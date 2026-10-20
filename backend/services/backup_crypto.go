package services

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// 数据备份加密包（T-072）：导出必须设置密码，对导出字节做密码加密。
//
// 文件格式：magic(6) | version(1) | iterations(4, big-endian) | salt(16) | nonce(12) | ciphertext
//   派生：PBKDF2-HMAC-SHA256(password, salt, iterations) → 32 字节密钥
//   加密：AES-256-GCM（密文尾部为 GCM tag）
// 导入兼容三种输入：加密包（须密码）/ 明文 xlsx（可无密码）/ 无密码加密包（报错要求密码）。

const (
	backupMagic      = "MCLBK1"
	backupVersion    = 1
	backupSaltLen    = 16
	backupNonceLen   = 12
	backupKeyLen     = 32
	backupIterations = 210000
)

var (
	// ErrBackupPasswordRequired 加密备份未提供密码（controller 映射 40001）
	ErrBackupPasswordRequired = errors.New("该文件为加密备份，请输入导出时设置的密码")
	// ErrBackupPasswordWrong 密码错误或文件被篡改（controller 映射 40001）
	ErrBackupPasswordWrong = errors.New("备份密码错误或文件已损坏")
	// ErrBackupPasswordEmpty 导出未设置密码（controller 映射 40001）
	ErrBackupPasswordEmpty = errors.New("导出必须设置密码")
)

// backupHeaderLen magic+version+iterations+salt+nonce
const backupHeaderLen = len(backupMagic) + 1 + 4 + backupSaltLen + backupNonceLen

// IsEncryptedBackup 判断字节是否为加密备份包（magic 头匹配）
func IsEncryptedBackup(data []byte) bool {
	return len(data) >= len(backupMagic) && string(data[:len(backupMagic)]) == backupMagic
}

// EncryptBackup 用密码加密导出字节，返回加密包；密码为空返回 ErrBackupPasswordEmpty
func EncryptBackup(plain []byte, password string) ([]byte, error) {
	if password == "" {
		return nil, ErrBackupPasswordEmpty
	}
	salt := make([]byte, backupSaltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	nonce := make([]byte, backupNonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	aead, err := backupAEAD(password, salt, backupIterations)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString(backupMagic)
	buf.WriteByte(backupVersion)
	if err := binary.Write(&buf, binary.BigEndian, uint32(backupIterations)); err != nil {
		return nil, err
	}
	buf.Write(salt)
	buf.Write(nonce)
	buf.Write(aead.Seal(nil, nonce, plain, backupAAD(backupIterations, salt)))
	return buf.Bytes(), nil
}

// DecryptBackup 用密码解开加密包；非加密包直接返回原字节（明文 xlsx 兼容）。
// 密码为空而文件已加密返回 ErrBackupPasswordRequired；密码错误返回 ErrBackupPasswordWrong。
func DecryptBackup(data []byte, password string) ([]byte, error) {
	if !IsEncryptedBackup(data) {
		return data, nil
	}
	if password == "" {
		return nil, ErrBackupPasswordRequired
	}
	if len(data) < backupHeaderLen+16 {
		return nil, errors.New("加密备份文件已损坏")
	}

	off := len(backupMagic)
	version := data[off]
	off++
	if version != backupVersion {
		return nil, errors.New("不支持的加密备份版本")
	}
	iterations := int(binary.BigEndian.Uint32(data[off : off+4]))
	off += 4
	if iterations <= 0 {
		return nil, errors.New("加密备份文件已损坏")
	}
	salt := data[off : off+backupSaltLen]
	off += backupSaltLen
	nonce := data[off : off+backupNonceLen]
	off += backupNonceLen
	ciphertext := data[off:]

	aead, err := backupAEAD(password, salt, iterations)
	if err != nil {
		return nil, err
	}
	plain, err := aead.Open(nil, nonce, ciphertext, backupAAD(iterations, salt))
	if err != nil {
		return nil, ErrBackupPasswordWrong
	}
	return plain, nil
}

// backupAAD 把版本/迭代次数/盐纳入认证范围，防头部被替换
func backupAAD(iterations int, salt []byte) []byte {
	aad := make([]byte, 0, len(backupMagic)+5+backupSaltLen)
	aad = append(aad, backupMagic...)
	aad = append(aad, backupVersion)
	var iter [4]byte
	binary.BigEndian.PutUint32(iter[:], uint32(iterations))
	aad = append(aad, iter[:]...)
	aad = append(aad, salt...)
	return aad
}

// backupAEAD 由密码 + 盐派生 AES-256 密钥并构造 GCM
func backupAEAD(password string, salt []byte, iterations int) (cipher.AEAD, error) {
	key := pbkdf2.Key([]byte(password), salt, iterations, backupKeyLen, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// EncryptBackupExt 加密包文件后缀（前端据此命名下载文件）
func EncryptBackupExt() string {
	return ".mcbak"
}
