#!/usr/bin/env python3
"""
TextProcessor Counter Script

文本统计脚本 - 用于执行高级文本统计功能。

作者: AgentFramework Team
版本: 1.2.0
许可证: MIT
"""

import sys
import re
import json
from typing import Dict, Any, List
from collections import Counter
import argparse


class TextCounter:
    """文本统计器"""

    def __init__(self):
        self.stop_words = {
            'a', 'an', 'the', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for',
            'of', 'with', 'by', 'from', 'as', 'is', 'was', 'are', 'were', 'be',
            'been', 'being', 'have', 'has', 'had', 'do', 'does', 'did', 'will',
            'would', 'could', 'should', 'may', 'might', 'must', 'can', 'this',
            'that', 'these', 'those', 'i', 'you', 'he', 'she', 'it', 'we', 'they'
        }

    def count_characters(self, text: str, include_spaces: bool = True) -> int:
        """统计字符数"""
        if include_spaces:
            return len(text)
        return len(text.replace(' ', '').replace('\n', '').replace('\t', ''))

    def count_words(self, text: str) -> int:
        """统计单词数"""
        words = re.findall(r'\b\w+\b', text.lower())
        return len(words)

    def count_lines(self, text: str) -> int:
        """统计行数"""
        return len(text.split('\n'))

    def count_paragraphs(self, text: str) -> int:
        """统计段落数"""
        paragraphs = [p.strip() for p in text.split('\n\n') if p.strip()]
        return len(paragraphs)

    def count_sentences(self, text: str) -> int:
        """统计句子数"""
        sentences = re.split(r'[.!?]+', text)
        sentences = [s.strip() for s in sentences if s.strip()]
        return len(sentences)

    def extract_keywords(self, text: str, limit: int = 10) -> List[str]:
        """提取关键词"""
        words = re.findall(r'\b\w+\b', text.lower())
        # 过滤停用词和短词
        words = [w for w in words if w not in self.stop_words and len(w) > 2]
        # 统计频率
        word_freq = Counter(words)
        # 返回最常见的词
        return [word for word, _ in word_freq.most_common(limit)]

    def estimate_reading_time(self, text: str) -> float:
        """估算阅读时间（分钟）"""
        words = self.count_words(text)
        # 假设平均阅读速度为 200 词/分钟
        return words / 200.0

    def analyze(self, text: str, options: Dict[str, Any] = None) -> Dict[str, Any]:
        """执行完整分析"""
        if options is None:
            options = {}

        result = {
            'characters': self.count_characters(text),
            'characters_no_spaces': self.count_characters(text, False),
            'words': self.count_words(text),
            'lines': self.count_lines(text),
            'paragraphs': self.count_paragraphs(text),
            'sentences': self.count_sentences(text),
            'reading_time': self.estimate_reading_time(text),
        }

        if options.get('extract_keywords', False):
            limit = options.get('keyword_limit', 10)
            result['keywords'] = self.extract_keywords(text, limit)

        if options.get('word_frequency', False):
            words = re.findall(r'\b\w+\b', text.lower())
            word_freq = Counter(words)
            result['word_frequency'] = dict(word_freq.most_common(20))

        return result


def main():
    """主函数"""
    parser = argparse.ArgumentParser(description='Text Counter - 文本统计工具')
    parser.add_argument('--input', '-i', required=True, help='输入文本或文件路径')
    parser.add_argument('--output', '-o', choices=['text', 'json'], default='text',
                      help='输出格式')
    parser.add_argument('--file', '-f', action='store_true', help='输入为文件路径')
    parser.add_argument('--keywords', '-k', action='store_true', help='提取关键词')
    parser.add_argument('--frequency', action='store_true', help='显示词频统计')
    parser.add_argument('--keyword-limit', type=int, default=10, help='关键词数量限制')

    args = parser.parse_args()

    # 读取输入
    try:
        if args.file:
            with open(args.input, 'r', encoding='utf-8') as f:
                text = f.read()
        else:
            text = args.input
    except FileNotFoundError:
        print(f"错误: 文件 '{args.input}' 不存在", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)

    # 执行分析
    counter = TextCounter()
    options = {
        'extract_keywords': args.keywords,
        'word_frequency': args.frequency,
        'keyword_limit': args.keyword_limit
    }
    result = counter.analyze(text, options)

    # 输出结果
    if args.output == 'json':
        print(json.dumps(result, indent=2, ensure_ascii=False))
    else:
        print("=" * 50)
        print("文本统计结果")
        print("=" * 50)
        print(f"字符数（含空格）: {result['characters']:,}")
        print(f"字符数（不含空格）: {result['characters_no_spaces']:,}")
        print(f"单词数: {result['words']:,}")
        print(f"行数: {result['lines']:,}")
        print(f"段落数: {result['paragraphs']:,}")
        print(f"句子数: {result['sentences']:,}")
        print(f"阅读时间: {result['reading_time']:.1f} 分钟")

        if 'keywords' in result:
            print("\n关键词:")
            for i, keyword in enumerate(result['keywords'], 1):
                print(f"  {i}. {keyword}")

        if 'word_frequency' in result:
            print("\n词频统计:")
            for word, freq in list(result['word_frequency'].items())[:10]:
                print(f"  {word}: {freq}")

    sys.exit(0)


if __name__ == '__main__':
    main()
