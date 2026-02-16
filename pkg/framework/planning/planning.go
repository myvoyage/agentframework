// Planning Package - 规划系统核心接口和类型
// Copyright (C) 2025 Agent Framework Contributors

package planning

import (
	"context"
	"fmt"
	"time"
)

// Step 规划步骤
type Step struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	Order             int               `json:"order"`
	Status            StepStatus        `json:"status"`
	Priority          int               `json:"priority"`
	EstimatedDuration time.Duration     `json:"estimated_duration"`
	Dependencies      []string          `json:"dependencies"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// StepStatus 步骤状态
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusInProgress StepStatus = "in_progress"
	StepStatusCompleted   StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
)

// PlanConfig 规划配置
type PlanConfig struct {
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
	MaxPlanningTime time.Duration `json:"max_planning_time"`
	EnableFallback  bool          `json:"enable_fallback"`
}

// Plan 规划
type Plan struct {
	ID                 string            `json:"id"`
	Timestamp          time.Time         `json:"timestamp"`
	Goal               string            `json:"goal"`
	Steps              []*Step           `json:"steps"`
	EstimatedDuration  time.Duration     `json:"estimated_duration"`
	Priority           int               `json:"priority"`
	Status             PlanStatus        `json:"status"`
	Dependencies       map[string][]string `json:"dependencies"`
	AssociatedContexts []string          `json:"associated_contexts,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// PlanStatus 规划状态
type PlanStatus string

const (
	PlanStatusDraft      PlanStatus = "draft"
	PlanStatusApproved   PlanStatus = "approved"
	PlanStatusExecuting  PlanStatus = "executing"
	PlanStatusCompleted  PlanStatus = "completed"
	PlanStatusFailed     PlanStatus = "failed"
	PlanStatusCancelled  PlanStatus = "cancelled"
)

// Validate 验证规划
func (p *Plan) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("plan ID cannot be empty")
	}
	if p.Goal == "" {
		return fmt.Errorf("plan goal cannot be empty")
	}
	if len(p.Steps) == 0 {
		return fmt.Errorf("plan must have at least one step")
	}
	return nil
}

// PlanGenerator 规划生成器接口
type PlanGenerator interface {
	Generate(ctx context.Context, goal string, context map[string]interface{}) (*Plan, error)
	Optimize(ctx context.Context, plan *Plan) (*Plan, error)
	AnalyzeFailure(ctx context.Context, plan *Plan, stepID string, err error) ([]string, error)
}

// BasePlanGenerator 基础规划生成器
type BasePlanGenerator struct {
	config *PlanConfig
}

func NewBasePlanGenerator(cfg *PlanConfig) *BasePlanGenerator {
	if cfg == nil {
		cfg = &PlanConfig{
			MaxRetries:      3,
			RetryDelay:      time.Second * 2,
			MaxPlanningTime: time.Minute * 5,
			EnableFallback:  true,
		}
	}
	return &BasePlanGenerator{
		config: cfg,
	}
}

func (bg *BasePlanGenerator) Generate(ctx context.Context, goal string, context map[string]interface{}) (*Plan, error) {
	// 基础实现，子类可以覆盖
	return &Plan{
		Goal:              goal,
		Status:            PlanStatusDraft,
	}, nil
}

func (bg *BasePlanGenerator) Optimize(ctx context.Context, plan *Plan) (*Plan, error) {
	return plan, nil
}

func (bg *BasePlanGenerator) AnalyzeFailure(ctx context.Context, plan *Plan, stepID string, err error) ([]string, error) {
	return []string{}, nil
}
