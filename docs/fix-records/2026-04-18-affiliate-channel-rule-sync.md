# 2026-04-18 推广链接渠道开关规则同步记录

## 1. 背景

本次整理的目标，不是新增一套独立业务，而是把已经落地的“推广链接渠道命中规则”正式固化为可追溯文档，避免后续继续按旧的三态理解误判页面权限。

本轮规则涉及以下核心页面与模块：

- `user/src/views/affiliate/TokenMerchantGuestV2.vue`
- `user/src/views/zhengye.vue`
- `user/src/views/Home.vue` 首页官方群栏目
- `admin` 端 Token 商配置入口

同时，本次文档同步基于 `MEMORY_BANK.md` 当前已确认事实进行整理，而不是重新假设业务。

## 2. 已确认的长期规则

### 2.1 全站基础三态规则继续保留

1. 游客不能访问 `TokenMerchantGuestV2.vue`
2. 游客不能访问 `zhengye.vue`
3. 注册客户默认不能访问 `zhengye.vue`
4. Token 商可以访问 `zhengye.vue`

### 2.2 新增“推广链接命中的渠道覆盖规则”

只有命中有效 `affiliate_code`，且该归因可解析到有效 Token 商推广上下文的访问者，才受该 Token 商自己的页面开关控制。

结论：

- 命中推广渠道：受对应商户配置影响
- 未命中推广渠道的自然流量：不受这两个渠道开关影响

### 2.3 `merchant_page_enabled` 的真实含义

该开关控制的是：

- 该 Token 商推广来的、已登录但尚未成为 Token 商的客户，是否还能进入 `TokenMerchantGuestV2.vue`

具体规则：

#### 开启时

- 命中该商户推广归因的注册客户，可进入 `TokenMerchantGuestV2.vue`
- 这些客户申请成功后，可进入 `zhengye.vue`

#### 关闭时

- 游客仍不可见
- 该商户推广来的已注册普通客户也不可见
- 已经成功开通过 Token 商的用户，**不回收** `zhengye.vue` 权限

### 2.4 `group_section_enabled` 的真实含义

该开关只控制：首页底部是否展示该 Token 商自己的官方群栏目。

具体规则：

- 只有命中该商户推广上下文的首页访问者，才会受此开关影响
- 开启：展示官方群栏目
- 关闭：隐藏官方群栏目
- 不影响其他商户渠道，也不影响自然流量首页

### 2.5 首页官方群栏目内容来源

当前内容直接复用 `affiliate_contacts`，不新建第二套配置表。

当前展示字段：

- `notice`
- `group_image_url`
- `parent_group_image_url`

## 3. 本轮已落实的前端行为

### 3.1 `user/src/router/index.ts`

- 对 `/token-merchant-v2` 增加推广渠道 + 开关校验
- 非 Token 商访问 `/zhengye` 时：
  - 若命中渠道且 `merchant_page_enabled=true`，跳转到 `/token-merchant-v2`
  - 否则跳回首页

### 3.2 `user/src/views/affiliate/TokenMerchantGuestV2.vue`

- 页面挂载时再次校验：
  - 登录态
  - Token 商身份
  - `affiliate_code`
  - `merchant_page_enabled`

### 3.3 `user/src/views/Home.vue`

- 命中渠道且 `group_section_enabled=true` 时，在首页底部展示官方群栏目

### 3.4 `user/src/components/Navbar.vue`

- 顶部自定义导航显式隐藏以下业务入口，避免直接暴露：
  - `/zhengye`
  - `/token-merchant-v2`

## 4. 本轮已落实的后台配套

### 4.1 `admin` 端配置语义

后台当前相关配置语义已统一为：

- `merchant_page_enabled`：控制命中该商户推广归因的注册客户是否还能进入招募页
- `group_section_enabled`：控制该商户渠道首页是否展示官方群栏目

### 4.2 后台菜单多语言同步

为避免规则落地后后台菜单仍出现错误文案，本轮同时校正：

- `promotionPlanSetting`
- `promotionStats`

三语文案：

- 简中：`推广方案设置` / `推广统计`
- 繁中：`推廣方案設定` / `推廣統計`
- 英文：`Promotion Plan Settings` / `Promotion Stats`

## 5. 验证结果

已确认：

- `user` 前端 `npm run build` 通过
- `admin` 前端 `npm run build` 通过
- 本轮文档整理与规则说明与 `MEMORY_BANK.md` 当前记录一致

## 6. 后续执行要求

后续凡是再处理以下问题，必须优先按本记录与 `MEMORY_BANK.md` 执行：

- `TokenMerchantGuestV2.vue` 可见性
- `zhengye.vue` 可见性
- 首页官方群栏目展示条件
- Token 商招募页开关语义
- Token 商渠道下的菜单、文案与配置解释

禁止回退为“只按游客 / 注册客户 / Token 商三态做静态判断、不看推广渠道上下文”的旧逻辑。