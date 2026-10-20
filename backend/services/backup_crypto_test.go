package services

import (
	"bytes"
	"errors"
	"testing"

	"mcloud/database"
	"mcloud/models"
	"mcloud/utils"

	"github.com/xuri/excelize/v2"
)

// TestBackupEncryptDecryptRoundtrip T-072：加密包带 magic 头，正确密码可还原原字节
func TestBackupEncryptDecryptRoundtrip(t *testing.T) {
	plain := []byte("sheet-bytes-\x00\x01\x02")
	enc, err := EncryptBackup(plain, "S3cret!pass")
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if !IsEncryptedBackup(enc) {
		t.Fatal("加密包缺少 magic 头")
	}
	if bytes.Contains(enc, plain) {
		t.Fatal("加密包内含明文")
	}
	got, err := DecryptBackup(enc, "S3cret!pass")
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("往返不一致: %q", got)
	}
}

// TestBackupEncryptRequiresPassword T-072：导出必须设置密码
func TestBackupEncryptRequiresPassword(t *testing.T) {
	if _, err := EncryptBackup([]byte("x"), ""); !errors.Is(err, ErrBackupPasswordEmpty) {
		t.Fatalf("空密码应返回 ErrBackupPasswordEmpty，实际 %v", err)
	}
}

// TestBackupDecryptErrors T-072：加密包无密码/错密码报错；明文 xlsx 无需密码直通
func TestBackupDecryptErrors(t *testing.T) {
	enc, err := EncryptBackup([]byte("payload"), "pw-123")
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if _, err := DecryptBackup(enc, ""); !errors.Is(err, ErrBackupPasswordRequired) {
		t.Fatalf("无密码应返回 ErrBackupPasswordRequired，实际 %v", err)
	}
	if _, err := DecryptBackup(enc, "wrong"); !errors.Is(err, ErrBackupPasswordWrong) {
		t.Fatalf("错密码应返回 ErrBackupPasswordWrong，实际 %v", err)
	}
	// 篡改密文尾部（GCM tag）
	tampered := bytes.Clone(enc)
	tampered[len(tampered)-1] ^= 0xff
	if _, err := DecryptBackup(tampered, "pw-123"); !errors.Is(err, ErrBackupPasswordWrong) {
		t.Fatalf("篡改应返回 ErrBackupPasswordWrong，实际 %v", err)
	}

	plain := []byte("plain-xlsx")
	if got, err := DecryptBackup(plain, ""); err != nil || !bytes.Equal(got, plain) {
		t.Fatalf("明文应直通: got=%q err=%v", got, err)
	}
	// 每次加密 salt/nonce 不同
	enc2, _ := EncryptBackup([]byte("payload"), "pw-123")
	if bytes.Equal(enc, enc2) {
		t.Fatal("两次加密结果不应相同（随机 salt/nonce）")
	}
}

// TestBackupEncryptedImport T-072：加密导出 → 空库凭密码导入成功；错密码报错；明文 xlsx 无密码导入成功
func TestBackupEncryptedImport(t *testing.T) {
	newTestDB(t)
	hs := NewHostService()
	req := baseHostReq("10.9.1.1")
	req.ObjectStorage = 800
	if _, err := hs.Create(req, testOp, NewRequestID()); err != nil {
		t.Fatalf("种子主机失败: %v", err)
	}

	data, err := NewDataExchangeService().Export()
	if err != nil {
		t.Fatalf("导出失败: %v", err)
	}
	enc, err := EncryptBackup(data, "backup-pw")
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	openTestDB(t, t.Name()+"_target")
	importer := NewDataExchangeService()

	_, err = importer.ImportWithPassword(enc, "", testOp, NewRequestID())
	if !errors.Is(err, ErrBackupPasswordRequired) {
		t.Fatalf("无密码导入应报 ErrBackupPasswordRequired，实际 %v", err)
	}
	_, err = importer.ImportWithPassword(enc, "bad", testOp, NewRequestID())
	if !errors.Is(err, ErrBackupPasswordWrong) {
		t.Fatalf("错密码导入应报 ErrBackupPasswordWrong，实际 %v", err)
	}

	report, err := importer.ImportWithPassword(enc, "backup-pw", testOp, NewRequestID())
	if err != nil {
		t.Fatalf("加密导入失败: %v", err)
	}
	if !report.Committed || len(report.Errors) != 0 {
		t.Fatalf("加密导入未成功: %+v", report.Errors)
	}
	var host models.Host
	if err = database.DB.Where("private_ip = ?", "10.9.1.1").First(&host).Error; err != nil {
		t.Fatalf("目标库缺主机: %v", err)
	}
	if host.ObjectStorage != 800 {
		t.Fatalf("object_storage=%d，期望 800", host.ObjectStorage)
	}

	// 明文 xlsx 直接导入（无密码）
	openTestDB(t, t.Name()+"_plain")
	report, err = NewDataExchangeService().ImportWithPassword(data, "", testOp, NewRequestID())
	if err != nil {
		t.Fatalf("明文导入失败: %v", err)
	}
	if !report.Committed {
		t.Fatalf("明文导入未提交: %+v", report.Errors)
	}
}

// TestCSVHeaderObjectStorage T-072：CSV 表头尾列为「对象存储(GB)」，27 列往返值不变
func TestCSVHeaderObjectStorage(t *testing.T) {
	if last := utils.CSVHeaders[len(utils.CSVHeaders)-1]; last != "对象存储(GB)" {
		t.Fatalf("CSV 末列表头=%q，期望 对象存储(GB)", last)
	}
	if bytes.Contains([]byte(stringsJoin(utils.CSVHeaders)), []byte("磁盘(GB)")) {
		t.Fatal("CSV 表头仍残留「磁盘(GB)」")
	}

	newTestDB(t)
	s := NewHostService()
	row := []string{
		"华东", "i-072", "csv-host", "10.9.2.1", "否",
		"虚拟机", "CentOS 7.9", "4", "X86", "8",
		"40", "100", "生产", "否",
		"运行中", "22", "web", "单位", "张三", "138", "项目",
		"理由", "配置", "2026-01-02", "50GB", "备注", "600",
	}
	req, err := s.ParseCSVRowToCreateHost(row)
	if err != nil {
		t.Fatalf("解析 CSV 行失败: %v", err)
	}
	if req.ObjectStorage != 600 {
		t.Fatalf("object_storage=%d，期望 600", req.ObjectStorage)
	}

	// 旧 26 列文件：无对象存储列，保持 0
	req, err = s.ParseCSVRowToCreateHost(row[:26])
	if err != nil {
		t.Fatalf("解析旧 26 列失败: %v", err)
	}
	if req.ObjectStorage != 0 {
		t.Fatalf("旧 26 列 object_storage=%d，期望 0", req.ObjectStorage)
	}
}

// TestXLSXObjectStorageColumn T-072：xlsx「主机」sheet 列头为「对象存储」，旧列头「磁盘」导入兼容
func TestXLSXObjectStorageColumn(t *testing.T) {
	for _, spec := range dataExchangeSheets {
		if spec.Name != "主机" {
			continue
		}
		found := false
		for i, h := range spec.Headers {
			if h == "对象存储" {
				found = true
				if i != 9 {
					t.Fatalf("对象存储列位置=%d，期望 9", i)
				}
			}
			if h == "磁盘" {
				t.Fatal("xlsx 主机 sheet 仍残留「磁盘」列头")
			}
		}
		if !found {
			t.Fatal("xlsx 主机 sheet 缺「对象存储」列头")
		}
	}

	// 旧备份（列头「磁盘」）也能导入
	newTestDB(t)
	hostHeaders := hostSheetHeaders("磁盘")
	data := buildWorkbookBytes(t, hostHeaders, []string{
		"华东", "i-074", "old-backup", "10.9.3.1", "虚拟机", "CentOS", "2", "X86", "4", "900",
		"40", "100", "生产", "否", "运行中", "22", "", "", "",
	})
	if _, err := NewDataExchangeService().ImportWithPassword(data, "", testOp, NewRequestID()); err != nil {
		t.Fatalf("旧列头导入失败: %v", err)
	}
	var host models.Host
	if err := database.DB.Where("private_ip = ?", "10.9.3.1").First(&host).Error; err != nil {
		t.Fatalf("旧列头导入缺主机: %v", err)
	}
	if host.ObjectStorage != 900 {
		t.Fatalf("旧列头 object_storage=%d，期望 900", host.ObjectStorage)
	}
}

func stringsJoin(ss []string) string {
	out := ""
	for _, s := range ss {
		out += s + ","
	}
	return out
}

// hostSheetHeaders 取主机 sheet 列头（可把「对象存储」替换为指定旧列头名）
func hostSheetHeaders(objectStorageHeader string) []string {
	for _, spec := range dataExchangeSheets {
		if spec.Name == "主机" {
			hs := append([]string{}, spec.Headers...)
			for i, h := range hs {
				if h == "对象存储" {
					hs[i] = objectStorageHeader
				}
			}
			return hs
		}
	}
	return nil
}

// buildWorkbookBytes 生成最小合法备份工作簿（8 sheet 齐全，仅主机 sheet 有 1 行数据）
func buildWorkbookBytes(t *testing.T, hostHeaders, hostRow []string) []byte {
	t.Helper()
	f := excelize.NewFile()
	defer f.Close()
	if err := f.SetSheetName("Sheet1", dataExchangeSheets[0].Name); err != nil {
		t.Fatalf("建表失败: %v", err)
	}
	for _, spec := range dataExchangeSheets[1:] {
		if _, err := f.NewSheet(spec.Name); err != nil {
			t.Fatalf("建表失败: %v", err)
		}
	}
	for _, spec := range dataExchangeSheets {
		if err := writeSheetRows(f, spec.Name, nil); err != nil {
			t.Fatalf("写 sheet %s 失败: %v", spec.Name, err)
		}
	}
	for i, h := range hostHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue("主机", cell, h); err != nil {
			t.Fatalf("写表头失败: %v", err)
		}
	}
	for i, v := range hostRow {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		if err := f.SetCellValue("主机", cell, v); err != nil {
			t.Fatalf("写主机行失败: %v", err)
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("生成工作簿失败: %v", err)
	}
	return buf.Bytes()
}
