package worker

import (
	"context"
	"errors"

	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/queue"

	"github.com/hibiken/asynq"
)

// Service 异步队列服务
type Service struct {
	name      string
	server    *asynq.Server
	mux       *asynq.ServeMux
	scheduler *asynq.Scheduler
	consumer  *Consumer
}

// NewService 创建异步队列服务
func NewService(cfg *config.QueueConfig, consumer *Consumer) (*Service, error) {
	if cfg == nil || !cfg.Enabled {
		return nil, errors.New("queue disabled")
	}
	if consumer == nil {
		return nil, errors.New("consumer is nil")
	}
	opt, serverCfg := queue.BuildServerConfig(cfg)
	server := asynq.NewServer(opt, serverCfg)
	mux := asynq.NewServeMux()
	consumer.Register(mux)

	scheduler := asynq.NewScheduler(opt, nil)
	registerPeriodicTasks(scheduler, consumer, cfg)

	return &Service{
		name:      "worker",
		server:    server,
		mux:       mux,
		scheduler: scheduler,
		consumer:  consumer,
	}, nil
}

// registerPeriodicTasks 注册所有周期性任务
func registerPeriodicTasks(scheduler *asynq.Scheduler, consumer *Consumer, cfg *config.QueueConfig) {
	if scheduler == nil || consumer == nil {
		return
	}
	if consumer.AffiliateService != nil {
		task := queue.NewAffiliateConfirmCommissionsTask()
		entryID, err := scheduler.Register("@every 1m", task, asynq.Queue(queue.DefaultQueue))
		if err != nil {
			logger.Warnw("scheduler_register_affiliate_confirm_failed", "error", err)
		} else {
			logger.Infow("scheduler_register_affiliate_confirm_ok", "entry_id", entryID)
		}
	}
	if consumer.LevelUpgradeService != nil {
		// 每天北京时间凌晨3点执行等级升级检查（CRON_TZ=Asia/Shanghai）
		task := queue.NewAffiliateLevelUpgradeCheckTask()
		entryID, err := scheduler.Register("CRON_TZ=Asia/Shanghai 0 3 * * *", task, asynq.Queue(queue.DefaultQueue))
		if err != nil {
			logger.Warnw("scheduler_register_affiliate_level_upgrade_failed", "error", err)
		} else {
			logger.Infow("scheduler_register_affiliate_level_upgrade_ok", "entry_id", entryID)
		}
	}
	if consumer.ProductMappingService != nil {
		syncInterval := "5m"
		if cfg != nil && cfg.UpstreamSyncInterval != "" {
			syncInterval = cfg.UpstreamSyncInterval
		}
		task := queue.NewUpstreamSyncStockTask()
		entryID, err := scheduler.Register("@every "+syncInterval, task, asynq.Queue(queue.DefaultQueue))
		if err != nil {
			logger.Warnw("scheduler_register_upstream_sync_stock_failed", "error", err)
		} else {
			logger.Infow("scheduler_register_upstream_sync_stock_ok", "entry_id", entryID, "interval", syncInterval)
		}
	}
	if consumer.NotificationService != nil {
		task, err := queue.NewNotificationInventoryAlertCheckTask()
		if err != nil {
			logger.Warnw("scheduler_register_inventory_alert_check_failed", "error", err)
		} else {
			entryID, registerErr := scheduler.Register("@every 1m", task, asynq.Queue(queue.DefaultQueue))
			if registerErr != nil {
				logger.Warnw("scheduler_register_inventory_alert_check_failed", "error", registerErr)
			} else {
				logger.Infow("scheduler_register_inventory_alert_check_ok", "entry_id", entryID)
			}
		}
	}
	if consumer.ProcurementOrderService != nil {
		task := queue.NewProcurementSyncAcceptedTask()
		entryID, err := scheduler.Register("@every 30m", task, asynq.Queue(queue.DefaultQueue))
		if err != nil {
			logger.Warnw("scheduler_register_procurement_sync_accepted_failed", "error", err)
		} else {
			logger.Infow("scheduler_register_procurement_sync_accepted_ok", "entry_id", entryID)
		}
	}
}

// Name 服务名称
func (s *Service) Name() string {
	if s == nil || s.name == "" {
		return "worker"
	}
	return s.name
}

// Start 启动服务
func (s *Service) Start(ctx context.Context) error {
	if s == nil || s.server == nil || s.mux == nil {
		return errors.New("worker not initialized")
	}
	if s.scheduler != nil {
		if err := s.scheduler.Start(); err != nil {
			logger.Warnw("scheduler_start_failed", "error", err)
		}
	}
	return s.server.Run(s.mux)
}

// Stop 停止服务
func (s *Service) Stop(ctx context.Context) error {
	if s == nil || s.server == nil {
		return nil
	}
	_ = ctx
	if s.scheduler != nil {
		s.scheduler.Shutdown()
	}
	s.server.Shutdown()
	return nil
}
