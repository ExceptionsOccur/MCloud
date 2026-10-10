#!/usr/bin/env bash
# MCloud 文档一致性自动校验（任务 T-021）
#
# 用法:
#   bash scripts/check_docs.sh                     # 全量检查
#   bash scripts/check_docs.sh --only links        # 只跑一项（links|tree|snapshot|counts|shas）
#   bash scripts/check_docs.sh --list              # 列出检查项
#
# 检查项:
#   links     全部已跟踪 *.md 的相对链接与锚点可解析（GitHub slug 规则，含 CJK 标题）
#   tree      docs/ARCHITECTURE.md 目录树 ↔ 真实文件（树中文件必须存在 error；已跟踪文件未登记 warn）
#   snapshot  AGENTS.md 状态快照 ↔ docs/ROADMAP.md（todo ID 集合、in_progress、next_task）
#   counts    计数类事实（N 个 SQL / N 张表 / N 个函数）与仓库实际一致
#   shas      ROADMAP done 表的提交号在 git 历史中存在
#
# 退出码: 0 = 无 error（warn 不计）; 1 = 存在 error; 2 = 用法错误
set -uo pipefail
export LC_ALL=C.UTF-8

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT" || exit 2

ONLY=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --only) ONLY="${2:-}"; shift 2 ;;
    --list) grep -E '^#   [a-z]+ ' "$0"; exit 0 ;;
    -h|--help) sed -n '2,16p' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
    *) printf '未知参数: %s\n' "$1" >&2; exit 2 ;;
  esac
done

ERR=0
WARN=0
err()  { printf 'ERROR %s\n' "$*"; ERR=$((ERR + 1)); }
warn() { printf 'WARN  %s\n' "$*"; WARN=$((WARN + 1)); }

md_files() { git ls-files -- '*.md'; }

# 标题 → GitHub 锚点：去 md 强调符 → 小写 → 删非「字母数字/空格/_/-」→ 空格转 -
slugify() {
  local s="$1"
  s="${s//\`/}"
  s="${s//\*\*}"
  s="${s//\*/}"
  printf '%s' "$s" | tr 'A-Z' 'a-z' | sed -E 's/[^[:alnum:] _-]//g; s/ /-/g'
}

# 输出某 md 文件的全部锚点（标题 slug + 显式 id）
anchors_for() {
  local f="$1" line
  [[ -f "$f" ]] || return 0
  while IFS= read -r line; do
    line="${line#\#* }"
    slugify "$line"
    printf '\n'
  done < <(grep -E '^#{1,6} ' "$f" || true)
  grep -oE "id=\"[^\"]+\"" "$f" | sed -E 's/id="([^"]+)"/\1/' || true
}

check_links() {
  local md lno content target path anchor resolved cache_dir cache
  cache_dir="$(mktemp -d)"

  for md in $(md_files); do
    while IFS= read -r content; do
      lno="${content%%:*}"
      content="${content#*:}"
      while IFS= read -r target; do
        [[ -z "$target" ]] && continue
        case "$target" in
          http://*|https://*|mailto:*) continue ;;
        esac
        target="${target%% *}"                       # 去掉可选 title
        path="${target%%#*}"
        anchor=""
        [[ "$target" == *'#'* ]] && anchor="${target#*#}"

        if [[ -z "$path" ]]; then
          resolved="$md"
        else
          resolved="$(realpath -m "$(dirname "$md")/$path")"
          resolved="${resolved#"$ROOT"/}"
          if [[ ! -f "$resolved" ]]; then
            err "$md:$lno [links] 链接目标不存在: $target"
            continue
          fi
        fi
        [[ -z "$anchor" ]] && continue
        cache="$cache_dir/${resolved//\//_}"
        [[ -f "$cache" ]] || anchors_for "$resolved" | sort -u > "$cache"
        if ! grep -qxF -- "$anchor" "$cache"; then
          err "$md:$lno [links] 锚点无法解析: $target"
        fi
      done < <(grep -oE '\]\([^)]+\)' <<<"$content" | sed -E 's/^\]\(//; s/\)$//')
    done < <(grep -nE '\]\([^)]+\)' "$md" || true)
  done
  rm -rf "$cache_dir"
}

check_tree() {
  local tree_file="docs/ARCHITECTURE.md"
  [[ -f "$tree_file" ]] || { err "[tree] 缺少 $tree_file"; return; }

  local tmp_dirs tmp_paths
  tmp_dirs="$(mktemp)"; tmp_paths="$(mktemp)"

  local line prefix name depth path dir i seen_root
  local -a stack=()
  seen_root=0

  while IFS= read -r line; do
    if [[ "$line" != *'── '* ]]; then
      # 代码块首行 `go/` 视为仓库根
      if [[ "$seen_root" -eq 0 && "$line" == go/ ]]; then
        seen_root=1
      fi
      continue
    fi
    prefix="${line%%── *}"
    name="${line#*── }"
    name="${name%% #*}"                               # 去行尾注释
    name="${name%"${name##*[![:space:]]}"}"           # 去尾部空白
    depth=$(( ${#prefix} / 4 ))                       # 每层 4 字符（│、├、└、空格均计 1 字符）

    path=""
    for ((i = 0; i < depth; i++)); do
      path+="${stack[i]}/"
    done

    if [[ "$name" == */ ]]; then
      dir="${path}${name%/}"
      if [[ ! -d "$dir" ]]; then
        err "$tree_file [tree] 目录树中的目录不存在: $dir"
      fi
      printf '%s\n' "$dir" >> "$tmp_dirs"
      stack[depth]="${name%/}"
    else
      if [[ ! -f "$path$name" ]]; then
        err "$tree_file [tree] 目录树中的文件不存在: $path$name"
      fi
      printf '%s\n' "$path$name" >> "$tmp_paths"
    fi
  done < <(awk '/^## 目录结构/{s=1; next} s==1 && /^```/{if(f==0){f=1; next}; exit} f==1{print}' "$tree_file")

  # 反向：已跟踪文件未登记进目录树
  # 规则：命中 → ok；否则找最近祖先目录，仅当该目录在树中「未展开」（无任何子文件登记，如 migrations/）才豁免
  local f ok anc
  for f in $(git ls-files -- backend frontend docker); do
    grep -qxF -- "$f" "$tmp_paths" && continue
    ok=0
    anc="$(dirname "$f")"
    while [[ "$anc" != "." && "$anc" != "/" ]]; do
      if grep -qxF -- "$anc" "$tmp_dirs"; then
        if ! grep -qE "^${anc}/" "$tmp_paths"; then
          ok=1   # 目录在树中但未展开，子文件豁免
        fi
        break
      fi
      anc="$(dirname "$anc")"
    done
    [[ "$ok" -eq 0 ]] && warn "$f:1 [tree] 已跟踪文件未登记到 ARCHITECTURE 目录树"
  done

  rm -f "$tmp_dirs" "$tmp_paths"
}

first_task_id() { grep -oE 'T-[0-9]{3}' <<<"$1" | head -1; }
ids_from() { grep -oE '\*\*T-[0-9]{3}\*\*' <<<"$1" | tr -d '*' | sort -u; }

check_snapshot() {
  local ag="AGENTS.md" roadmap="docs/ROADMAP.md"
  local ag_todo rm_todo ag_ip rm_ip ag_next rm_first diff_out

  if [[ ! -f "$ag" || ! -f "$roadmap" ]]; then
    err "[snapshot] 缺少 $ag 或 $roadmap"; return
  fi

  ag_todo="$(grep -E '^\| todo（P[12]） \|' "$ag" | grep -oE 'T-[0-9]{3}' | sort -u)"
  rm_todo="$(ids_from "$(sed -n '/^## todo/,/^## blocked/p' "$roadmap")")"
  diff_out="$(comm -3 <(printf '%s\n' "$ag_todo" | sed '/^$/d') <(printf '%s\n' "$rm_todo" | sed '/^$/d'))"
  if [[ -n "$diff_out" ]]; then
    err "[snapshot] AGENTS 快照 todo 与 ROADMAP 不一致（首列=仅AGENTS，次列=仅ROADMAP）:"
    while IFS= read -r l; do printf '       %s\n' "$l"; done <<<"$diff_out"
  fi

  ag_ip="$(grep -E '^\| in_progress \|' "$ag" | grep -oE 'T-[0-9]{3}' | sort -u)"
  rm_ip="$(ids_from "$(sed -n '/^## in_progress/,/^## todo/p' "$roadmap")")"
  if [[ "$ag_ip" != "$rm_ip" ]]; then
    warn "[snapshot] AGENTS in_progress 与 ROADMAP 不一致（进行中属预期，回写协议第4步须刷新）: AGENTS=[$(tr '\n' ',' <<<"$ag_ip")] ROADMAP=[$(tr '\n' ',' <<<"$rm_ip")]"
  fi

  ag_next="$(grep -E '^\| next_task \|' "$ag" || true)"
  rm_first="$(grep -oE '\*\*T-[0-9]{3}\*\*' <<<"$(sed -n '/^## todo/,/^## blocked/p' "$roadmap")" | head -1 | tr -d '*')"
  if [[ -z "$ag_next" ]]; then
    warn "[snapshot] AGENTS 状态快照缺少 next_task 字段（应等于 ROADMAP todo 首条: ${rm_first:-无}）"
  else
    local next_id
    next_id="$(first_task_id "$ag_next")"
    if [[ -z "$next_id" ]]; then
      err "[snapshot] AGENTS next_task 字段无法解析出 T-xxx"
    elif [[ "$next_id" != "$rm_first" ]]; then
      err "[snapshot] next_task=$next_id 与 ROADMAP todo 首条 ${rm_first:-无} 不一致"
    fi
  fi

  local lu
  lu="$(grep -E '^\| last_updated \|' "$ag" | grep -oE '[0-9]{4}-[0-9]{2}-[0-9]{2}' | head -1)"
  [[ -z "$lu" ]] && warn "[snapshot] AGENTS last_updated 缺失或不是 YYYY-MM-DD"
}

# 校验计数类描述: $1=文件 $2=行号 $3=描述文本(如 '8 个 SQL') $4=实际值
check_count_stmt() {
  local f="$1" lno="$2" text="$3" actual="$4"
  local claimed="${text%% *}"
  if [[ "$claimed" != "$actual" ]]; then
    err "$f:$lno [counts] 「$text」与实际 $actual 不一致"
  fi
}

check_counts() {
  local f lno text line actual gofile

  # 1) N 个 SQL ↔ backend/migrations/*.sql
  actual="$(ls backend/migrations/*.sql 2>/dev/null | wc -l | tr -d ' ')"
  while IFS=: read -r f lno text; do
    check_count_stmt "$f" "$lno" "$text" "$actual"
  done < <(grep -rnE '[0-9]+ 个( `[^`]*`)? SQL' docs README.md AGENTS.md |
    sed -E 's/^([^:]+):([0-9]+):.*[^0-9]([0-9]+ 个( `[^`]*`)? SQL).*$/\1:\2:\3/')

  # 2) N 张表 / N 张业务表 / N 张： ↔ ARCHITECTURE 数据模型章节数
  actual="$(sed -n '/^## 数据模型/,$p' docs/ARCHITECTURE.md | grep -cE '^### .* 表')"
  while IFS=: read -r f lno text; do
    check_count_stmt "$f" "$lno" "$text" "$actual"
  done < <(grep -rnE '[0-9]+ 张(业务表|表|：)' docs README.md AGENTS.md |
    sed -E 's/^([^:]+):([0-9]+):.*[^0-9]([0-9]+ 张(业务表|表|：)).*$/\1:\2:\3/')

  # 3) N 个函数 ↔ 行内引用的 .go 文件 func 数（行内无文件路径则跳过）
  while IFS=: read -r f lno line; do
    gofile="$(grep -oE '[a-zA-Z0-9_/.-]+\.go' <<<"$line" | head -1)"
    [[ -z "$gofile" ]] && continue
    [[ "$gofile" == backend/* ]] || gofile="backend/$gofile"
    [[ -f "$gofile" ]] || continue
    actual="$(grep -cE '^func ' "$gofile")"
    text="$(grep -oE '[0-9]+ 个函数' <<<"$line" | head -1)"
    [[ -n "$text" ]] && check_count_stmt "$f" "$lno" "$text" "$actual"
  done < <(grep -rnE '[0-9]+ 个函数' docs README.md AGENTS.md)
}

check_shas() {
  local line sha
  while IFS= read -r line; do
    while IFS= read -r sha; do
      [[ -z "$sha" ]] && continue
      if ! git cat-file -e "${sha}^{commit}" 2>/dev/null; then
        err "docs/ROADMAP.md [shas] done 表提交号不存在: $sha"
      fi
    done < <(grep -oE '`[0-9a-f]{7,40}`' <<<"$line" | tr -d '`')
  done < <(sed -n '/^## done/,$p' docs/ROADMAP.md | grep -E '^\|')
}

run() {
  local name="$1"
  if [[ -n "$ONLY" && "$ONLY" != "$name" ]]; then return 0; fi
  "check_$name"
}

if [[ -n "$ONLY" && ! "$ONLY" =~ ^(links|tree|snapshot|counts|shas)$ ]]; then
  printf '未知检查项: %s\n' "$ONLY" >&2
  exit 2
fi

run links
run tree
run snapshot
run counts
run shas

printf '\ncheck_docs: %d errors, %d warnings\n' "$ERR" "$WARN"
[[ "$ERR" -eq 0 ]] || exit 1
exit 0
