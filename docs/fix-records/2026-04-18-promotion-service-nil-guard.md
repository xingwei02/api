# 2026-04-18 promotion 服务空值兜底修复记录

## 1. 背景

本次排查围绕游客订单预览 / 商品优惠展示链路展开。

结合 `MEMORY_BANK.md` 的执行原则，先核实真实代码结构，再落修复，避免按错误路径盲改。

## 2. 排查结论

### 2.1 已确认的事实

1. `zhengye` 推广中心的客户优惠配置（`affiliate_discounts`）**当前未接入订单金额计算链**。
2. 当前仓库中**不存在**以下用户猜测路径：
   - `api/internal/service/promotion/calculator.go`
   - `CalculatePromotion`
3. 当前订单优惠计算真实链路为：
   - `internal/service/order_service_validate.go`
   - `internal/service/promotion_service.go`
   - `internal/service/member_level_service.go`
4. `PromotionService.ApplyPromotion()` 当前实现为**实时查库**：
   - `promotionRepo.GetAllActiveByProduct(product.ID, time.Now())`
   - 未发现“服务启动时缓存档位列表”的实现证据。
5. `OrderService` 的 `promotionRepo` 在 `internal/provider/container.go` 中存在注入，同时 `internal/service/order_service.go` 已确认构造函数中保留：
   - `promotionRepo: opts.PromotionRepo`

### 2.2 本次修复目标

虽然未证实“缓存档位”问题，但为了防止真实运行中因为依赖为空导致预览/展示链路异常，本次对 promotion 服务补充**空值兜底保护**：

- 当 `PromotionService` 为 `nil` 时，不崩溃
- 当 `promotionRepo` 为 `nil` 时，不崩溃
- 商品活动价查询失败前，先保证空 repo 不会进入查询
- 回退策略统一为：
  - 展示类接口返回空活动列表
  - 金额计算类接口回退为商品原价

## 3. 实际修改内容

### 3.1 `internal/service/promotion_service.go`

#### `GetProductPromotions(productID uint)`

新增保护：

- `s == nil`
- `s.promotionRepo == nil`
- `productID == 0`

处理方式：直接返回空数组 `[]models.Promotion{}`。

#### `ApplyPromotion(product *models.Product, quantity int)`

在原有参数校验后新增保护：

- `s == nil`
- `s.promotionRepo == nil`

处理方式：直接返回：

- `promotion = nil`
- `price = product.PriceAmount`
- `error = nil`

即：**安全降级到原价**。

### 3.2 测试文件

新增：`internal/service/promotion_service_test.go`

覆盖两个场景：

1. `ApplyPromotion()` 在 repo 为 nil 时优雅回退原价
2. `GetProductPromotions()` 在 repo 为 nil 时返回空列表

## 4. 验证结果

已执行：

```bash
go test ./internal/service -run TestPromotionService -count=1
```

结果：

```text
ok github.com/dujiao-next/internal/service 0.038s
```

## 5. 涉及文件

- `internal/service/order_service.go`
- `internal/service/promotion_service.go`
- `internal/service/promotion_service_test.go`

## 6. 备注

本次是**防御式稳定性修复**，不是对一个实际存在的“promotion calculator 缓存档位模块”进行改造。

如果后续线上仍出现游客预览 500，需要继续结合真实响应体 / 运行日志做二次定位，但本次至少已经把 promotion 服务的空 repo 场景封住。