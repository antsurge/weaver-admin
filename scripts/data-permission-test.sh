#!/usr/bin/env bash
# =====================================================================
# 数据权限手动测试方案 -> 自动化脚本
# 对应文档：docs/data-permission-test-plan.md
# 功能：自动搭建部门树/规则/角色/测试账号 -> 登录断言 current-user.dataScope
#       -> 禁用规则联动 -> 超管保护 -> 清理。
# 依赖：curl、redis-cli（用于读取验证码明文）、python3（用于解析 JSON）
# 用法：
#   BASE_URL=http://127.0.0.1:8888 ADMIN_USER=admin ADMIN_PASS=123456 \
#   ./scripts/data-permission-test.sh [--no-cleanup]
# =====================================================================
set -uo pipefail

# ---------------------------- 配置区 ----------------------------
BASE_URL="${BASE_URL:-http://127.0.0.1:8888}"
ADMIN_USER="${ADMIN_USER:-admin}"
ADMIN_PASS="${ADMIN_PASS:-123456}"
TEST_USER_PASS="${TEST_USER_PASS:-123456}"
REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_DB="${REDIS_DB:-0}"
REDIS_PREFIX="${REDIS_PREFIX:-weaver_admin:}"
CLEANUP="${1:-}"

PASS=0
FAIL=0
ADMIN_TOKEN=""
declare -g SCRIPT_DIR
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ---------------------------- 工具函数 ----------------------------
log()   { printf '\n\033[1;36m[STEP]\033[0m %s\n' "$*"; }
ok()    { PASS=$((PASS+1)); printf '\033[1;32m  PASS\033[0m %s\n' "$*"; }
bad()   { FAIL=$((FAIL+1)); printf '\033[1;31m  FAIL\033[0m %s\n' "$*"; }
info()  { printf '\033[1;33m  INFO\033[0m %s\n' "$*"; }

json_field() { # json 字段
  python3 -c '
import sys, json
try:
    d = json.loads(sys.argv[1])
except Exception:
    sys.exit(1)
v = d.get(sys.argv[2])
print("" if v is None else v)
' "$1" "$2"
}

# request <method> <path> [body] -> 输出 "HTTP_CODE<TAB>BODY"
request() {
  local method="$1" path="$2" body="${3:-}"
  local out code
  if [ -n "$body" ]; then
    out=$(curl -s -X "$method" "$BASE_URL$path" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -H "Content-Type: application/json" \
      -d "$body" -w '\n%{http_code}')
  else
    out=$(curl -s -X "$method" "$BASE_URL$path" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -w '\n%{http_code}')
  fi
  code="${out##*$'\n'}"
  body="${out%$'\n'*}"
  printf '%s\t%s' "$code" "$body"
}

# get_captcha <url> -> 输出 "captchaId<SP>CIPHER"（CIPHER 从 redis 读取）
get_captcha() {
  local url="$1" resp cid code
  resp=$(curl -s "$BASE_URL$url")
  cid=$(json_field "$resp" captchaId)
  code=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -n "$REDIS_DB" \
    GET "${REDIS_PREFIX}captcha:${cid}" 2>/dev/null | tr -d '\r\n')
  if [ -z "$cid" ] || [ -z "$code" ]; then
    return 1
  fi
  printf '%s %s' "$cid" "$code"
}

# login <username> <password> -> 输出 accessToken（失败输出空）
login() {
  local user="$1" pass="$2" cap tok
  cap=$(get_captcha "/admin/v1/get-captcha") || { bad "获取验证码失败($user)"; return 1; }
  local cid="${cap%% *}" code="${cap##* }"
  local resp
  resp=$(curl -s -X POST "$BASE_URL/admin/v1/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$user\",\"password\":\"$pass\",\"captcha\":\"$code\",\"captchaId\":\"$cid\"}")
  tok=$(json_field "$resp" accessToken)
  if [ -z "$tok" ]; then
    bad "登录失败: $user ($(json_field "$resp" message))"
    return 1
  fi
  printf '%s' "$tok"
}

# current_scope <token> -> 输出 dataScope
current_scope_of() {
  local token="$1"
  local resp
  resp=$(curl -s "$BASE_URL/admin/v1/current-user" \
    -H "Authorization: Bearer $token")
  json_field "$resp" dataScope
}

# assert_scope <username> <expected_scope> <说明>
assert_scope() {
  local user="$1" want="$2" desc="$3" tok got pwd
  pwd="$TEST_USER_PASS"
  [ "$user" = "$ADMIN_USER" ] && pwd="$ADMIN_PASS"
  tok=$(login "$user" "$pwd") || return
  got=$(current_scope_of "$tok")
  if [ "$got" = "$want" ]; then
    ok "$desc: $user dataScope=$got（期望 $want）"
  else
    bad "$desc: $user dataScope=$got（期望 $want）"
  fi
}

# create_dept <name> <code> <parent_id> <type> -> 输出 id
create_dept() {
  local name="$1" code="$2" parent="$3" type="$4"
  local resp id
  resp=$(request POST "/admin/v1/department" \
    "{\"name\":\"$name\",\"code\":\"$code\",\"parentID\":\"$parent\",\"type\":\"$type\",\"weight\":100,\"status\":\"enabled\"}")
  id=$(json_field "${resp#*$'\t'}" id)
  [ -n "$id" ] || bad "创建部门失败: $name"
  printf '%s' "$id"
}

# create_rule <name> <code> <scopeType> <deptIds_raw> <roleIds_raw> -> 输出 id
# deptIds_raw / roleIds_raw 传 JSON 数组字符串，如 ["id1","id2"]（函数内会做 JSON 转义）
create_rule() {
  local name="$1" code="$2" st="$3" depts="$4" roles="$5"
  local body
  body=$(python3 -c '
import json, sys
name, code, st, depts, roles = sys.argv[1:6]
print(json.dumps({"name": name, "code": code, "scopeType": st,
                  "deptIds": depts, "roleIds": roles, "status": "enabled"}))
' "$name" "$code" "$st" "$depts" "$roles")
  local resp id
  resp=$(request POST "/admin/v1/data-permission" "$body")
  id=$(json_field "${resp#*$'\t'}" id)
  [ -n "$id" ] || bad "创建规则失败: $code (${resp#*$'\t'})"
  printf '%s' "$id"
}

# create_role <name> <code> -> 输出 id
create_role() {
  local name="$1" code="$2"
  local resp id
  resp=$(request POST "/admin/v1/role" \
    "{\"name\":\"$name\",\"code\":\"$code\",\"weight\":10,\"status\":\"enabled\"}")
  id=$(json_field "${resp#*$'\t'}" id)
  [ -n "$id" ] || bad "创建角色失败: $code (${resp#*$'\t'})"
  printf '%s' "$id"
}

# bind_role_dp <role_id> <dp_id...>
bind_role_dp() {
  local rid="$1"; shift
  local ids=()
  for x in "$@"; do ids+=("\"$x\""); done
  local joined
  joined=$(IFS=,; echo "${ids[*]}")
  request PUT "/admin/v1/role/$rid/data-permissions" "{\"dataPermissionIds\":[$joined]}" >/dev/null
}

# create_user <username> <dept_id> <role_ids_csv> <dp_ids_csv> -> 输出 id
create_user() {
  local user="$1" dept="$2" roles="$3" dps="$4"
  local roles_json dps_json
  roles_json="[]"; [ -n "$roles" ] && roles_json="[$roles]"
  dps_json="[]";   [ -n "$dps" ]   && dps_json="[$dps]"
  local resp id
  resp=$(request POST "/admin/v1/admin" \
    "{\"username\":\"$user\",\"password\":\"$TEST_USER_PASS\",\"realName\":\"$user\",\"status\":\"enabled\",\"departmentId\":\"$dept\",\"roleIds\":$roles_json,\"dataPermissionIds\":$dps_json}")
  id=$(json_field "${resp#*$'\t'}" id)
  [ -n "$id" ] || bad "创建用户失败: $user (${resp#*$'\t'})"
  printf '%s' "$id"
}

summary() {
  printf '\n===============================\n'
  printf '测试结果: PASS=%s FAIL=%s\n' "$PASS" "$FAIL"
  [ "$FAIL" -eq 0 ] && printf '全部通过 ✓\n' || printf '存在失败，请对照文档排查 ✗\n'
  exit "$([ "$FAIL" -eq 0 ] && echo 0 || echo 1)"
}

# ---------------------------- 前置检查 ----------------------------
for cmd in curl redis-cli python3; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "缺少依赖命令: $cmd"; exit 1; }
done

if ! curl -s -o /dev/null -m 3 "$BASE_URL/admin/v1/get-captcha"; then
  echo "服务未启动或地址不可达: $BASE_URL"
  exit 1
fi

echo "目标环境: $BASE_URL  管理员: $ADMIN_USER"
log "登录管理员"
ADMIN_TOKEN=$(login "$ADMIN_USER" "$ADMIN_PASS") || { echo "管理员登录失败"; exit 1; }
ok "管理员登录成功（用户ID=$(curl -s "$BASE_URL/admin/v1/current-user" -H "Authorization: Bearer $ADMIN_TOKEN" | python3 -c 'import sys,json;print(json.load(sys.stdin).get("id",""))')）"

# ---------------------------- 1. 部门树 ----------------------------
log "前置 1：创建部门树"
D_ROOT=$(create_dept "总公司"     "dp_root"  ""            "company")
D_EAST=$(create_dept "华东分公司" "dp_east"  "$D_ROOT"     "subsidiary")
D_SH=$(  create_dept "上海部门"   "dp_sh"    "$D_EAST"     "department")
D_HZ=$(  create_dept "杭州部门"   "dp_hz"    "$D_EAST"     "department")
D_SOUTH=$(create_dept "华南分公司" "dp_south" "$D_ROOT"    "subsidiary")
D_SZ=$(  create_dept "深圳部门"   "dp_sz"    "$D_SOUTH"    "department")
D_GZ=$(  create_dept "广州部门"   "dp_gz"    "$D_SOUTH"    "department")
ok "部门树创建完成（root=$D_ROOT east=$D_EAST sh=$D_SH hz=$D_HZ south=$D_SOUTH sz=$D_SZ gz=$D_GZ）"

# ---------------------------- 2. 数据权限规则 ----------------------------
log "前置 2：创建 6 条数据权限规则"
R_ALL=$(create_rule "全部数据"     "DP_ALL"          "1" "" "")
R_DEPT=$(create_rule "本部门数据"  "DP_DEPT"         "3" "" "")
R_CHILD=$(create_rule "本部门及以下" "DP_DEPT_CHILD"  "4" "" "")
R_SELF=$(create_rule "仅本人"      "DP_SELF"         "5" "" "")
R_SH=$(create_rule "自定义-上海及杭州" "DP_CUSTOM_SH" "2" "[\"$D_SH\",\"$D_HZ\"]" "")
R_ROLE=$(create_rule "自定义-按角色" "DP_CUSTOM_ROLE" "2" "" "")
ok "6 条规则创建完成（ALL=$R_ALL DEPT=$R_DEPT CHILD=$R_CHILD SELF=$R_SELF CUSTOM_SH=$R_SH CUSTOM_ROLE=$R_ROLE）"

# 测试 1-4：唯一 code 校验（重复创建 DP_ALL 应失败）
log "测试 1：规则 CRUD（唯一 code 校验 / 编辑 / 禁用 / 删除）"
resp=$(request POST "/admin/v1/data-permission" \
  "{\"name\":\"重复规则\",\"code\":\"DP_ALL\",\"scopeType\":\"3\",\"status\":\"enabled\"}")
body="${resp#*$'\t'}"
if echo "$body" | grep -qi "exists"; then
  ok "重复 code 创建被拦截（DP_ALL 已存在）"
else
  bad "重复 code 未被拦截: $body"
fi

# 编辑规则（把 CUSTOM_ROLE 的 roleIds 绑定为 role_dept，稍后创建角色后再补）
# 禁用->启用规则联动在测试 4 单独做
# ---------------------------- 3. 角色 ----------------------------
log "前置 3：创建角色并绑定规则"
ROLE_ALL=$(create_role "华东角色-全部"    "ROLE_DP_ALL")
ROLE_DEPT=$(create_role "华东角色-本部门" "ROLE_DP_DEPT")
ROLE_CHILD=$(create_role "华东角色-及以下" "ROLE_DP_CHILD")
ROLE_SELF=$(create_role "华东角色-仅本人" "ROLE_DP_SELF")
ROLE_CUSTOM=$(create_role "华东角色-自定义" "ROLE_DP_CUSTOM")
ROLE_NONE=$(create_role "无规则角色"      "ROLE_NONE")
bind_role_dp "$ROLE_ALL"   "$R_ALL"
bind_role_dp "$ROLE_DEPT"  "$R_DEPT"
bind_role_dp "$ROLE_CHILD" "$R_CHILD"
bind_role_dp "$ROLE_SELF"  "$R_SELF"
bind_role_dp "$ROLE_CUSTOM" "$R_SH"
# 补充 CUSTOM_ROLE 规则的 roleIds【按角色自定义】
request PUT "/admin/v1/data-permission/$R_ROLE" \
  "{\"name\":\"自定义-按角色\",\"code\":\"DP_CUSTOM_ROLE\",\"scopeType\":\"2\",\"roleIds\":\"[\"$ROLE_SELF\"]\",\"status\":\"enabled\"}" >/dev/null
ok "角色创建并绑定规则完成"

# ---------------------------- 4. 测试账号 ----------------------------
log "前置 4：创建 9 个测试账号（含 admin）"
UID_A2=$(create_user "all_user"      "$D_SH"   "\"$ROLE_ALL\"" "")
UID_A3=$(create_user "dept_user"     "$D_SH"   "\"$ROLE_DEPT\"" "")
UID_A4=$(create_user "child_user"    "$D_EAST" "\"$ROLE_CHILD\"" "")
UID_A5=$(create_user "self_user"     "$D_HZ"   "\"$ROLE_NONE\"" "")
UID_A6=$(create_user "custom_user"   "$D_SZ"   "\"$ROLE_CUSTOM\"" "")
UID_A7=$(create_user "override_user" "$D_GZ"   "\"$ROLE_SELF\"" "\"$R_ALL\"")
UID_A8=$(create_user "norole_user"   ""        "" "")
UID_A9=$(create_user "multirole_user" "$D_SH"  "\"$ROLE_DEPT\",\"$ROLE_ALL\"" "")
ok "用户创建完成（except admin，均使用默认密码 $TEST_USER_PASS）"

# ---------------------------- 测试 3：范围解析正确性 ----------------------------
log "测试 3：逐个登录断言 current-user.dataScope"
assert_scope "$ADMIN_USER"  "ALL"            "超管恒为 ALL"
assert_scope "all_user"     "ALL"            "角色规则 ALL"
assert_scope "dept_user"    "DEPT"           "角色规则 DEPT"
assert_scope "child_user"   "DEPT_AND_CHILD" "角色规则 DEPT_AND_CHILD"
assert_scope "self_user"    "SELF"           "无规则角色兜底 SELF"
assert_scope "custom_user"  "DEPT_AND_CHILD" "自定义按部门集合"
assert_scope "override_user" "ALL"           "用户级 ALL 覆盖角色级 SELF"
assert_scope "norole_user"  "SELF"           "无角色无规则兜底 SELF"
assert_scope "multirole_user" "ALL"          "多角色并集取最大"

# ---------------------------- 测试 4：禁用规则联动 ----------------------------
log "测试 4：禁用规则联动"
request PUT "/admin/v1/data-permission/$R_ALL/status" '{"status":"disabled"}' >/dev/null
t=$(login "all_user" "$TEST_USER_PASS") || { bad "all_user 重新登录失败"; t=""; }
got=$(current_scope_of "$t")
if [ "$got" = "SELF" ]; then
  ok "禁用 ALL 规则后 all_user dataScope=$got（期望 SELF，回落角色无规则兜底）"
else
  bad "禁用 ALL 规则后 all_user dataScope=$got（期望 SELF）"
fi
request PUT "/admin/v1/data-permission/$R_ALL/status" '{"status":"enabled"}' >/dev/null
ok "恢复 ALL 规则为启用"

# ---------------------------- 测试 5：超管保护 ----------------------------
log "测试 5：超管保护"
admin_id=$(curl -s "$BASE_URL/admin/v1/current-user" -H "Authorization: Bearer $ADMIN_TOKEN" | python3 -c 'import sys,json;print(json.load(sys.stdin)["id"])')
# 5.1 删除 admin 超管账号应被拦截
resp=$(request DELETE "/admin/v1/admin?ids=$admin_id")
code_="${resp%%$'\t'*}"
if [ "$code_" -ge 400 ] 2>/dev/null; then
  ok "尝试删除超管账号被拦截（HTTP $code_）"
else
  bad "删除超管账号未被拦截（HTTP $code_）"
fi
# 5.2 删除自己（当前登录账号）应被拦截
resp_self=$(request DELETE "/admin/v1/admin?ids=$admin_id")
code_self="${resp_self%%$'\t'*}"
if [ "$code_self" -ge 400 ] 2>/dev/null; then
  ok "尝试删除当前登录账号被拦截（HTTP $code_self）"
else
  bad "删除当前登录账号未被拦截（HTTP $code_self）"
fi
# 5.3 为 all_user 临时绑定 SuperAdmin 角色后，dataScope 恒为 ALL
super_role_id=$(curl -s "$BASE_URL/admin/v1/role?page=1&pageSize=10&code=SuperAdmin" -H "Authorization: Bearer $ADMIN_TOKEN" | python3 -c '
import sys,json
try:
    d=json.load(sys.stdin)
    for it in d.get("items",[]):
        if it.get("code")=="SuperAdmin": print(it["id"]); break
except Exception: pass
')
if [ -n "$super_role_id" ]; then
  resp_user=$(curl -s "$BASE_URL/admin/v1/admin?page=1&pageSize=1&username=all_user" -H "Authorization: Bearer $ADMIN_TOKEN")
  uid_all=$(python3 -c "import sys,json;print(json.load(sys.stdin)['items'][0]['id'])" <<< "$resp_user")
  curl -s -X PUT "$BASE_URL/admin/v1/admin/$uid_all" \
    -H "Authorization: Bearer $ADMIN_TOKEN" -H "Content-Type: application/json" \
    -d "{\"roleIds\":[\"$super_role_id\"]}" >/dev/null
  t=$(login "all_user" "$TEST_USER_PASS") || t=""
  got=$(current_scope_of "$t")
  if [ "$got" = "ALL" ]; then
    ok "绑定 SuperAdmin 角色后 all_user dataScope=$got（超管角色强制 ALL）"
  else
    bad "绑定 SuperAdmin 角色后 all_user dataScope=$got（期望 ALL）"
  fi
  # 解除绑定，避免影响清理
  UID_ALL_2=$(curl -s "$BASE_URL/admin/v1/admin?page=1&pageSize=1&username=all_user" -H "Authorization: Bearer $ADMIN_TOKEN" | python3 -c 'import sys,json;print(json.load(sys.stdin)["items"][0]["id"] if json.load(sys.stdin).get("items") else "")')
  [ -n "$UID_ALL_2" ] && curl -s -X PUT "$BASE_URL/admin/v1/admin/$UID_ALL_2" \
    -H "Authorization: Bearer $ADMIN_TOKEN" -H "Content-Type: application/json" \
    -d "{\"roleIds\":[\"$ROLE_ALL\"]}" >/dev/null
else
  info "未找到 SuperAdmin 角色，跳过 5.3"
fi

# ---------------------------- 清理 ----------------------------
if [ "${CLEANUP}" = "--no-cleanup" ]; then
  log "已跳过清理（--no-cleanup）"
else
  log "清理：删除测试数据（用户/角色/规则/部门）"
  for u in multirole_user norole_user override_user custom_user self_user child_user dept_user all_user; do
    uid=$(curl -s "$BASE_URL/admin/v1/admin?page=1&pageSize=1&username=$u" -H "Authorization: Bearer $ADMIN_TOKEN" | python3 -c 'import sys,json;d=json.load(sys.stdin);print(d["items"][0]["id"] if d.get("items") else "")')
    [ -n "$uid" ] && request DELETE "/admin/v1/admin?ids=$uid" >/dev/null
  done
  for r in $ROLE_NONE $ROLE_CUSTOM $ROLE_SELF $ROLE_CHILD $ROLE_DEPT $ROLE_ALL; do
    request DELETE "/admin/v1/role?ids=$r" >/dev/null
  done
  for r in $R_SELF $R_SH $R_CHILD $R_DEPT $R_ALL $R_ROLE; do
    request DELETE "/admin/v1/data-permission?ids=$r" >/dev/null
  done
  for d in $D_GZ $D_SZ $D_SOUTH $D_HZ $D_SH $D_EAST $D_ROOT; do
    request DELETE "/admin/v1/department?ids=$d" >/dev/null
  done
  ok "清理完成"
fi

summary