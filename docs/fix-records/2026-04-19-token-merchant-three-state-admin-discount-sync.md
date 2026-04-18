# 2026-04-19 Token 商三态与 admin 折扣语义修正记录

## 背景

本轮不是新增业务，而是修正已经落地实现中与既定规则不一致的部分。

问题来源：

1. 注册客户访问 `TokenMerchantGuestV2.vue` 的逻辑被错误收紧为“必须命中 affiliate code 且 `merchant_page_enabled=true`”
2. admin 端 Token 商折扣配置错误落到了非目标语义字段，且范围与文档不一致

## 本轮确认后的长期规则

### 1. 三态访问规则

固定前提如下：

1. 游客
   - 不能访问 `TokenMerchantGuestV2.vue`
   - 不能访问 `zhengye.vue`
2. 注册客户
   - 可以访问 `TokenMerchantGuestV2.vue`
   - 不能访问 `zhengye.vue`
   - 可在宣传页申请成为 Token 商
3. Token 商
   - 可以访问 `TokenMerchantGuestV2.vue`
   - 可以访问 `zhengye.vue`

### 2. `merchant_page_enabled` 的真实作用域

该字段不是全站总开关，也不是 Token 商身份总权限开关。

它只控制：

- **命中某个 Token 商推广归因的注册客户**，能否继续看到该 Token 商对应的 `TokenMerchantGuestV2.vue` 宣传页

因此：

- 普通自然流量注册客户，不应因为没有 `affiliate_code` 被误拦截
- 已经成为 Token 商的用户，不因该字段关闭而回收 `zhengye.vue` 权限

### 3. admin 端折扣配置唯一落点

Token 商客户优惠配置统一使用：

- `affiliate_discounts`

本轮确认字段语义：

- `discount_rate`
- `merchant_page_enabled`
- `group_section_enabled`

`discount_rate` 当前业务范围固定为：

- `0 ~ 5`

未经再次确认，不应把这套 Token 商客户优惠配置重新写回其他平行表，也不应再拆第二套重复语义。

## 本轮实际修正

### user 端

- `/token-merchant-v2` 改为：注册客户可访问，游客不可访问
- 非 Token 商访问 `/zhengye` 时，统一跳到 `/token-merchant-v2`
- `TokenMerchantGuestV2.vue` 页面挂载时：
  - 游客强制返回
  - 已是 Token 商保留后续进入正业中心能力
  - 仅当命中推广归因时，才继续检查 `merchant_page_enabled`

### api 端

- `GET /admin/affiliates/users/:id/discount`
- `PUT /admin/affiliates/users/:id/discount`

统一改为读取 / 写入 `affiliate_discounts`，返回字段包含：

- `discount_rate`
- `merchant_page_enabled`
- `group_section_enabled`

并将 `discount_rate` 校验修正为 `0 ~ 5`。

### admin 端

- `AffiliateUsers.vue` 的“拿货折扣”操作已同步改为编辑：
  - `discount_rate`
  - `merchant_page_enabled`
  - `group_section_enabled`
- 不再只更新单个折扣值

## 验证结果

- `api`: `go build ./...` 通过
- `user`: `npm run build` 通过
- `admin`: `npm run build` 通过

## 后续排查建议

以后再遇到 Token 商相关问题，默认按下面顺序确认：

1. 当前用户身份：游客 / 注册客户 / Token 商
2. 当前访问页面：宣传页还是正业页
3. 是否命中有效推广归因
4. `affiliate_discounts` 中三个字段是否与页面表现一致
5. admin 与 user 两侧保存 / 展示语义是否一致