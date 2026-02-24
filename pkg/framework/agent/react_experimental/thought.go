//go:build experimental

// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package react

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"AgentFramework/pkg/framework/agent"
	"AgentFramework/pkg/framework/memory"
	"AgentFramework/pkg/errors"
)

// ThoughtProcessor 处理思考步骤的处理器接口
// 【必须】定义思考处理逻辑的标准接口
type ThoughtProcessor interface {
	// Process 处理思考步骤
	// 【必须】接收上下文和思考对象，返回处理后的思考和可能的错误
	Process(ctx context.Context, thought *Thought, state *ReActState) (*Thought, error)
	// Validate 验证思考处理器配置
	// 【必须】验证处理器自身的配置有效性
	Validate() error
	// Name 返回处理器名称
	// 【必须】提供处理器标识
	Name() string
}

// BaseThoughtProcessor 思考处理器的基类实现
// 【推荐】提供基础实现减少重复代码
type BaseThoughtProcessor struct {
	name   string
	logger *zap.Logger
}

// NewBaseThoughtProcessor 创建基础思考处理器
// 【必须】提供构造函数确保必要字段初始化
func NewBaseThoughtProcessor(name string, logger *zap.Logger) *BaseThoughtProcessor {
	if logger == nil {
		// 【必须】提供默认日志记录器避免nil指针
		logger = zap.NewNop()
	}

	return &BaseThoughtProcessor{
		name:   name,
		logger: logger.With(zap.String("processor", name)),
	}
}

// Name 返回处理器名称
// 【必须】实现 ThoughtProcessor 接口的 Name 方法
func (btp *BaseThoughtProcessor) Name() string {
	return btp.name
}

// Validate 验证处理器配置
// 【必须】实现 ThoughtProcessor 接口的 Validate 方法
func (btp *BaseThoughtProcessor) Validate() error {
	if btp.name == "" {
		return errors.NewValidationError("processor name cannot be empty", nil)
	}
	return nil
}

// ReasoningEnhancer 推理增强处理器
// 【必须】专门处理推理过程的增强和优化
type ReasoningEnhancer struct {
	BaseThoughtProcessor
	// MinConfidence 最小置信度阈值
	MinConfidence float64
	// MaxReasoningLength 最大推理长度
	MaxReasoningLength int
	// EnableValidation 是否启用推理验证
	EnableValidation bool
}

// NewReasoningEnhancer 创建推理增强处理器
// 【必须】提供构造函数确保必要配置
func NewReasoningEnhancer(logger *zap.Logger, minConfidence float64, maxLength int, enableValidation bool) *ReasoningEnhancer {
	processor := &ReasoningEnhancer{
		BaseThoughtProcessor: *NewBaseThoughtProcessor("reasoning_enhancer", logger),
		MinConfidence:        minConfidence,
		MaxReasoningLength:   maxLength,
		EnableValidation:     enableValidation,
	}

	// 【必须】设置合理的默认值
	if processor.MinConfidence < 0.0 || processor.MinConfidence > 1.0 {
		processor.MinConfidence = 0.3
	}

	if processor.MaxReasoningLength <= 0 {
		processor.MaxReasoningLength = 1000
	}

	return processor
}

// Process 处理思考步骤，增强推理过程
// 【必须】实现 ThoughtProcessor 接口的 Process 方法
func (re *ReasoningEnhancer) Process(ctx context.Context, thought *Thought, state *ReActState) (*Thought, error) {
	if thought == nil {
		return nil, errors.NewValidationError("thought cannot be nil", nil)
	}

	if state == nil {
		return nil, errors.NewValidationError("state cannot be nil", nil)
	}

	// 【必须】记录处理开始
	re.logger.Debug("starting reasoning enhancement",
		zap.String("thought_id", thought.ID),
		zap.Float64("original_confidence", thought.Confidence),
		zap.Int("reasoning_length", len(thought.Reasoning)),
	)

	// 创建思考副本避免修改原对象
	processedThought := thought.Clone()

	// 置信度检查和调整
	if processedThought.Confidence < re.MinConfidence {
		re.logger.Warn("thought confidence below threshold, enhancing",
			zap.Float64("current_confidence", processedThought.Confidence),
			zap.Float64("threshold", re.MinConfidence),
		)

		processedThought.Confidence = re.enhanceConfidence(ctx, processedThought, state)
	}

	// 推理内容增强
	if processedThought.Reasoning != "" {
		enhancedReasoning, err := re.enhanceReasoning(ctx, processedThought, state)
		if err != nil {
			re.logger.Warn("failed to enhance reasoning, using original",
				zap.Error(err),
				zap.String("thought_id", processedThought.ID),
			)
		} else {
			processedThought.Reasoning = enhancedReasoning
		}
	} else {
		// 生成基础推理
		basicReasoning, err := re.generateBasicReasoning(ctx, processedThought, state)
		if err != nil {
			re.logger.Warn("failed to generate basic reasoning",
				zap.Error(err),
				zap.String("thought_id", processedThought.ID),
			)
		} else {
			processedThought.Reasoning = basicReasoning
		}
	}

	// 推理长度控制
	if len(processedThought.Reasoning) > re.MaxReasoningLength {
		re.logger.Info("reasoning exceeds maximum length, truncating",
			zap.Int("original_length", len(processedThought.Reasoning)),
			zap.Int("max_length", re.MaxReasoningLength),
		)
		processedThought.Reasoning = processedThought.Reasoning[:re.MaxReasoningLength]
	}

	// 推理验证
	if re.EnableValidation {
		if err := re.validateReasoning(ctx, processedThought); err != nil {
			re.logger.Warn("reasoning validation failed",
				zap.Error(err),
				zap.String("thought_id", processedThought.ID),
			)
			// 【必须】验证失败不应阻止处理，只记录警告
		}
	}

	// 更新时间戳
	processedThought.Timestamp = time.Now().UTC()

	// 【必须】记录处理完成
	re.logger.Debug("reasoning enhancement completed",
		zap.String("thought_id", processedThought.ID),
		zap.Float64("final_confidence", processedThought.Confidence),
		zap.Int("final_reasoning_length", len(processedThought.Reasoning)),
	)

	return processedThought, nil
}

// enhanceConfidence 增强思考置信度
// 【推荐】提供置信度提升逻辑
func (re *ReasoningEnhancer) enhanceConfidence(ctx context.Context, thought *Thought, state *ReActState) float64 {
	// 基于历史表现调整置信度
	baseAdjustment := 0.2

	// 检查历史成功率
	if state.Observations != nil && len(state.Observations) > 0 {
		successCount := 0
		for _, obs := range state.Observations {
			if obs.Success {
				successCount++
			}
		}

		successRate := float64(successCount) / float64(len(state.Observations))
		baseAdjustment += successRate * 0.3
	}

	// 基于思考内容质量调整
	if len(thought.Content) > 50 {
		baseAdjustment += 0.1
	}

	newConfidence := thought.Confidence + baseAdjustment
	if newConfidence > 1.0 {
		newConfidence = 1.0
	}

	return newConfidence
}

// enhanceReasoning 增强推理内容
// 【推荐】提供推理内容优化逻辑
func (re *ReasoningEnhancer) enhanceReasoning(ctx context.Context, thought *Thought, state *ReActState) (string, error) {
	// 基于上下文信息丰富推理
	enhancedReasoning := thought.Reasoning

	// 添加历史上下文
	if state.Thoughts != nil && len(state.Thoughts) > 0 {
		recentThought := state.Thoughts[len(state.Thoughts)-1]
		enhancedReasoning += fmt.Sprintf("\n[基于前次思考: %s]", recentThought.Content[:min(len(recentThought.Content), 100)])
	}

	// 添加目标导向
	enhancedReasoning += fmt.Sprintf("\n[目标导向: %s]", state.Query)

	// 添加置信度说明
	enhancedReasoning += fmt.Sprintf("\n[置信度: %.2f, 基于%d次迭代经验]", thought.Confidence, state.IterationCount)

	return enhancedReasoning, nil
}

// generateBasicReasoning 生成基础推理
// 【推荐】当缺乏推理内容时生成基础推理
func (re *ReasoningEnhancer) generateBasicReasoning(ctx context.Context, thought *Thought, state *ReActState) (string, error) {
	basicReasoning := fmt.Sprintf("基于当前目标'%s'和思考内容进行分析:", state.Query)
	basicReasoning += fmt.Sprintf("\n- 思考要点: %s", thought.Content)
	basicReasoning += fmt.Sprintf("\n- 当前置信度: %.2f", thought.Confidence)
	basicReasoning += fmt.Sprintf("\n- 迭代次数: %d", state.IterationCount)

	// 如果有记忆，添加相关信息
	if state.Memory != nil {
		// 【必须】这里可以添加记忆检索逻辑
		basicReasoning += "\n- 参考历史经验和知识"
	}

	return basicReasoning, nil
}

// validateReasoning 验证推理内容
// 【推荐】提供推理质量验证
func (re *ReasoningEnhancer) validateReasoning(ctx context.Context, thought *Thought) error {
	if thought.Reasoning == "" {
		return errors.NewValidationError("reasoning cannot be empty after enhancement", nil)
	}

	if len(thought.Reasoning) < 10 {
		return errors.NewValidationError(
			"reasoning too short, may lack sufficient analysis",
			map[string]interface{}{"reasoning_length": len(thought.Reasoning)},
		)
	}

	// 检查推理是否包含逻辑连接词
	logicalConnectors := []string{"因为", "所以", "因此", "基于", "考虑到", "由于"}
	foundConnector := false
	for _, connector := range logicalConnectors {
		if thought.Reasoning == connector {
			foundConnector = true
			break
		}
	}

	if !foundConnector {
		re.logger.Debug("reasoning may lack logical connectors",
			zap.String("thought_id", thought.ID),
		)
	}

	return nil
}

// Validate 验证推理增强处理器配置
// 【必须】实现 ThoughtProcessor 接口的 Validate 方法
func (re *ReasoningEnhancer) Validate() error {
	if err := re.BaseThoughtProcessor.Validate(); err != nil {
		return errors.WrapError(err, "base processor validation failed", nil)
	}

	if re.MinConfidence < 0.0 || re.MinConfidence > 1.0 {
		return errors.NewValidationError(
			"minimum confidence must be between 0.0 and 1.0",
			map[string]interface{}{"min_confidence": re.MinConfidence},
		)
	}

	if re.MaxReasoningLength <= 0 {
		return errors.NewValidationError(
			"maximum reasoning length must be positive",
			map[string]interface{}{"max_length": re.MaxReasoningLength},
		)
	}

	return nil
}

// ContextAnalyzer 上下文分析处理器
// 【必须】分析思考步骤的上下文关联性
type ContextAnalyzer struct {
	BaseThoughtProcessor
	// AnalysisDepth 分析深度
	AnalysisDepth int
	// EnableCrossReference 是否启用交叉引用
	EnableCrossReference bool
	// MemoryManager 记忆管理器
	MemoryManager memory.Manager
}

// NewContextAnalyzer 创建上下文分析处理器
// 【必须】提供构造函数确保必要配置
func NewContextAnalyzer(logger *zap.Logger, analysisDepth int, enableCrossRef bool, memoryMgr memory.Manager) *ContextAnalyzer {
	processor := &ContextAnalyzer{
		BaseThoughtProcessor: *NewBaseThoughtProcessor("context_analyzer", logger),
		AnalysisDepth:        analysisDepth,
		EnableCrossReference: enableCrossRef,
		MemoryManager:        memoryMgr,
	}

	// 【必须】设置合理的默认值
	if processor.AnalysisDepth <= 0 {
		processor.AnalysisDepth = 3
	}

	if processor.MemoryManager == nil {
		processor.logger.Warn("memory manager is nil, context analysis will be limited")
	}

	return processor
}

// Process 处理思考步骤，分析上下文关联性
// 【必须】实现 ThoughtProcessor 接口的 Process 方法
func (ca *ContextAnalyzer) Process(ctx context.Context, thought *Thought, state *ReActState) (*Thought, error) {
	if thought == nil {
		return nil, errors.NewValidationError("thought cannot be nil", nil)
	}

	// 【必须】记录处理开始
	ca.logger.Debug("starting context analysis",
		zap.String("thought_id", thought.ID),
		zap.Int("analysis_depth", ca.AnalysisDepth),
	)

	// 创建思考副本避免修改原对象
	processedThought := thought.Clone()

	// 分析上下文关联性
	contextAnalysis, err := ca.analyzeContext(ctx, processedThought, state)
	if err != nil {
		ca.logger.Warn("context analysis failed",
			zap.Error(err),
			zap.String("thought_id", processedThought.ID),
		)
		// 【必须】分析失败不应阻止处理
	} else {
		// 将分析结果添加到思考上下文中
		processedThought.Context["context_analysis"] = contextAnalysis
	}

	// 交叉引用分析
	if ca.EnableCrossReference {
		crossRefs, err := ca.findCrossReferences(ctx, processedThought, state)
		if err != nil {
			ca.logger.Warn("cross-reference analysis failed",
				zap.Error(err),
				zap.String("thought_id", processedThought.ID),
			)
		} else {
			processedThought.Context["cross_references"] = crossRefs
		}
	}

	// 更新时间戳
	processedThought.Timestamp = time.Now().UTC()

	// 【必须】记录处理完成
	ca.logger.Debug("context analysis completed",
		zap.String("thought_id", processedThought.ID),
		zap.Bool("has_analysis", processedThought.Context["context_analysis"] != nil),
	)

	return processedThought, nil
}

// analyzeContext 分析思考上下文关联性
// 【推荐】提供上下文分析逻辑
func (ca *ContextAnalyzer) analyzeContext(ctx context.Context, thought *Thought, state *ReActState) (map[string]interface{}, error) {
	analysis := make(map[string]interface{})

	// 分析历史思考关联性
	if state.Thoughts != nil && len(state.Thoughts) > 0 {
		recentThoughts := ca.getRecentThoughts(state, ca.AnalysisDepth)
		thoughtSimilarities := ca.calculateThoughtSimilarities(thought, recentThoughts)
		analysis["thought_similarities"] = thoughtSimilarities
		analysis["historical_context_strength"] = ca.calculateContextStrength(thoughtSimilarities)
	}

	// 分析目标关联性
	goalAlignment := ca.analyzeGoalAlignment(thought, state.Query)
	analysis["goal_alignment"] = goalAlignment

	// 分析动作关联性
	if state.Actions != nil && len(state.Actions) > 0 {
		actionRelevance := ca.analyzeActionRelevance(thought, state.Actions)
		analysis["action_relevance"] = actionRelevance
	}

	return analysis, nil
}

// findCrossReferences 查找交叉引用
// 【推荐】提供交叉引用分析逻辑
func (ca *ContextAnalyzer) findCrossReferences(ctx context.Context, thought *Thought, state *ReActState) (map[string]interface{}, error) {
	crossRefs := make(map[string]interface{})

	// 如果有记忆管理器，从记忆中查找相关信息
	if ca.MemoryManager != nil {
		// 【必须】这里可以添加记忆检索逻辑
		// 示例：检索与当前思考相关的历史信息
		relatedMemories, err := ca.retrieveRelatedMemories(ctx, thought.Content)
		if err != nil {
			ca.logger.Debug("failed to retrieve related memories", zap.Error(err))
		} else {
			crossRefs["related_memories"] = relatedMemories
		}
	}

	// 查找相似的历史思考
	similarThoughts := ca.findSimilarThoughts(thought, state.Thoughts)
	if len(similarThoughts) > 0 {
		crossRefs["similar_thoughts"] = similarThoughts
	}

	return crossRefs, nil
}

// getRecentThoughts 获取最近的思考
// 【推荐】辅助方法获取最近的思考记录
func (ca *ContextAnalyzer) getRecentThoughts(state *ReActState, depth int) []*Thought {
	if state.Thoughts == nil {
		return []*Thought{}
	}

	start := 0
	if len(state.Thoughts) > depth {
		start = len(state.Thoughts) - depth
	}

	return state.Thoughts[start:]
}

// calculateThoughtSimilarities 计算思考相似度
// 【推荐】计算思考之间的相似度
func (ca *ContextAnalyzer) calculateThoughtSimilarities(current *Thought, historical []*Thought) map[string]float64 {
	similarities := make(map[string]float64)

	for i, historicalThought := range historical {
		similarity := ca.calculateSimilarity(current.Content, historicalThought.Content)
		similarities[historicalThought.ID] = similarity
	}

	return similarities
}

// calculateSimilarity 计算文本相似度（简单实现）
// 【推荐】提供基础的文本相似度计算
func (ca *ContextAnalyzer) calculateSimilarity(text1, text2 string) float64 {
	// 简单的词汇重叠度计算
	words1 := ca.extractWords(text1)
	words2 := ca.extractWords(text2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	commonWords := 0
	wordSet := make(map[string]bool)

	for _, word := range words1 {
		wordSet[word] = true
	}

	for _, word := range words2 {
		if wordSet[word] {
			commonWords++
		}
	}

	// 计算 Jaccard 相似度
	totalUniqueWords := len(wordSet)
	for _, word := range words2 {
		wordSet[word] = true
	}
	totalUniqueWords = len(wordSet)

	if totalUniqueWords == 0 {
		return 0.0
	}

	return float64(commonWords) / float64(totalUniqueWords)
}

// extractWords 提取词汇
// 【推荐】辅助方法提取文本词汇
func (ca *ContextAnalyzer) extractWords(text string) []string {
	// 简单的分词实现（实际应用中应该使用更复杂的分词器）
	words := make([]string, 0)
	currentWord := ""

	for _, char := range text {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			currentWord += string(char)
		} else {
			if len(currentWord) > 1 {
				words = append(words, currentWord)
			}
			currentWord = ""
		}
	}

	if len(currentWord) > 1 {
		words = append(words, currentWord)
	}

	return words
}

// calculateContextStrength 计算上下文强度
// 【推荐】基于相似度计算上下文强度
func (ca *ContextAnalyzer) calculateContextStrength(similarities map[string]float64) float64 {
	if len(similarities) == 0 {
		return 0.0
	}

	totalSimilarity := 0.0
	for _, similarity := range similarities {
		totalSimilarity += similarity
	}

	return totalSimilarity / float64(len(similarities))
}

// analyzeGoalAlignment 分析目标对齐度
// 【推荐】分析思考与目标的对齐程度
func (ca *ContextAnalyzer) analyzeGoalAlignment(thought *Thought, goal string) float64 {
	// 简单的目标关键词匹配
	goalWords := ca.extractWords(goal)
	thoughtWords := ca.extractWords(thought.Content)

	if len(goalWords) == 0 || len(thoughtWords) == 0 {
		return 0.0
	}

	goalMatches := 0
	for _, goalWord := range goalWords {
		for _, thoughtWord := range thoughtWords {
			if goalWord == thoughtWord {
				goalMatches++
				break
			}
		}
	}

	return float64(goalMatches) / float64(len(goalWords))
}

// analyzeActionRelevance 分析动作相关性
// 【推荐】分析思考与历史动作的相关性
func (ca *ContextAnalyzer) analyzeActionRelevance(thought *Thought, actions []*Action) map[string]float64 {
	relevance := make(map[string]float64)

	for _, action := range actions {
		// 基于动作名称和思考内容的简单匹配
		matchScore := ca.calculateActionThoughtMatch(action, thought)
		relevance[action.ID] = matchScore
	}

	return relevance
}

// calculateActionThoughtMatch 计算动作与思考的匹配度
// 【推荐】计算动作与思考的匹配分数
func (ca *ContextAnalyzer) calculateActionThoughtMatch(action *Action, thought *Thought) float64 {
	// 简单的文本匹配
	actionText := action.Name + " " + action.Description
	thoughtText := thought.Content

	return ca.calculateSimilarity(actionText, thoughtText)
}

// retrieveRelatedMemories 检索相关记忆
// 【推荐】从记忆中检索相关信息
func (ca *ContextAnalyzer) retrieveRelatedMemories(ctx context.Context, content string) ([]memory.MemoryItem, error) {
	if ca.MemoryManager == nil {
		return nil, errors.NewValidationError("memory manager is nil", nil)
	}

	// 【必须】这里应该调用记忆管理器的检索方法
	// 由于具体的记忆管理器接口可能不同，这里提供示例实现
	// relatedMemories, err := ca.MemoryManager.RetrieveRelevant(ctx, content, 5)

	// 示例返回空数组
	return []memory.MemoryItem{}, nil
}

// findSimilarThoughts 查找相似的思考
// 【推荐】在历史思考中查找相似项
func (ca *ContextAnalyzer) findSimilarThoughts(target *Thought, thoughts []*Thought) []*Thought {
	similar := make([]*Thought, 0)
	threshold := 0.3 // 相似度阈值

	for _, thought := range thoughts {
		similarity := ca.calculateSimilarity(target.Content, thought.Content)
		if similarity >= threshold {
			similar = append(similar, thought)
		}
	}

	return similar
}

// Validate 验证上下文分析处理器配置
// 【必须】实现 ThoughtProcessor 接口的 Validate 方法
func (ca *ContextAnalyzer) Validate() error {
	if err := ca.BaseThoughtProcessor.Validate(); err != nil {
		return errors.WrapError(err, "base processor validation failed", nil)
	}

	if ca.AnalysisDepth <= 0 {
		return errors.NewValidationError(
			"analysis depth must be positive",
			map[string]interface{}{"analysis_depth": ca.AnalysisDepth},
		)
	}

	return nil
}

// ThoughtChain 思考链管理器
// 【必须】管理思考步骤的执行链
type ThoughtChain struct {
	// processors 思考处理器链
	processors []ThoughtProcessor
	// logger 日志记录器
	logger *zap.Logger
}

// NewThoughtChain 创建思考链管理器
// 【必须】提供构造函数确保必要初始化
func NewThoughtChain(logger *zap.Logger) *ThoughtChain {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &ThoughtChain{
		processors: make([]ThoughtProcessor, 0),
		logger:     logger.Named("thought_chain"),
	}
}

// AddProcessor 添加思考处理器
// 【必须】提供处理器管理功能
func (tc *ThoughtChain) AddProcessor(processor ThoughtProcessor) error {
	if processor == nil {
		return errors.NewValidationError("processor cannot be nil", nil)
	}

	if err := processor.Validate(); err != nil {
		return errors.WrapError(err, "processor validation failed", nil)
	}

	tc.processors = append(tc.processors, processor)
	tc.logger.Debug("added thought processor",
		zap.String("processor_name", processor.Name()),
		zap.Int("chain_length", len(tc.processors)),
	)

	return nil
}

// Process 按顺序处理思考步骤
// 【必须】实现思考链的完整处理流程
func (tc *ThoughtChain) Process(ctx context.Context, thought *Thought, state *ReActState) (*Thought, error) {
	if thought == nil {
		return nil, errors.NewValidationError("thought cannot be nil", nil)
	}

	if state == nil {
		return nil, errors.NewValidationError("state cannot be nil", nil)
	}

	// 【必须】记录处理开始
	tc.logger.Debug("starting thought chain processing",
		zap.String("thought_id", thought.ID),
		zap.Int("processors_count", len(tc.processors)),
	)

	currentThought := thought
	var lastError error

	// 按顺序执行所有处理器
	for i, processor := range tc.processors {
		select {
		case <-ctx.Done():
			return nil, errors.WrapError(ctx.Err(), "thought chain processing cancelled", nil)
		default:
			{
				processedThought, err := processor.Process(ctx, currentThought, state)
				if err != nil {
					tc.logger.Error("processor failed in thought chain",
						zap.Error(err),
						zap.String("processor_name", processor.Name()),
						zap.Int("processor_index", i),
						zap.String("thought_id", currentThought.ID),
					)
					lastError = errors.WrapError(err, "processor failed in chain", map[string]interface{}{
						"processor_name":  processor.Name(),
						"processor_index": i,
						"thought_id":      currentThought.ID,
					})
					// 【必须】处理器失败不应停止整个链，继续下一个处理器
					continue
				}

				if processedThought != nil {
					currentThought = processedThought
				}
			}
		}
	}

	// 【必须】记录处理完成
	tc.logger.Debug("thought chain processing completed",
		zap.String("original_thought_id", thought.ID),
		zap.String("final_thought_id", currentThought.ID),
		zap.Error(lastError),
	)

	return currentThought, lastError
}

// Clear 清空处理器链
// 【推荐】提供链重置功能
func (tc *ThoughtChain) Clear() {
	tc.processors = make([]ThoughtProcessor, 0)
	tc.logger.Debug("cleared thought processor chain")
}

// ProcessorCount 返回处理器数量
// 【推荐】提供链状态查询
func (tc *ThoughtChain) ProcessorCount() int {
	return len(tc.processors)
}

// Validate 验证思考链配置
// 【必须】验证整个链的完整性
func (tc *ThoughtChain) Validate() error {
	if tc.processors == nil {
		return errors.NewValidationError("processors slice cannot be nil", nil)
	}

	for i, processor := range tc.processors {
		if processor == nil {
			return errors.NewValidationError(
				fmt.Sprintf("processor at index %d cannot be nil", i),
				map[string]interface{}{"index": i},
			)
		}

		if err := processor.Validate(); err != nil {
			return errors.WrapError(err, fmt.Sprintf("processor at index %d validation failed", i), nil)
		}
	}

	return nil
}

// Helper function for min calculation
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}