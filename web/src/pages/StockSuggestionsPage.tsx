import { DeepVoidBackground } from '../components/DeepVoidBackground'
import { Container } from '../components/Container'
import { useLanguage } from '../contexts/LanguageContext'

export function StockSuggestionsPage() {
  const { language } = useLanguage()

  return (
    <DeepVoidBackground className="min-h-screen py-12">
      <Container>
        <div className="flex flex-col items-center justify-center text-center space-y-6">
          <div className="w-24 h-24 rounded-full bg-blue-500/10 flex items-center justify-center border border-blue-500/30 animate-pulse">
            <span className="text-4xl">🤖</span>
          </div>
          <h1 className="text-4xl font-bold text-nofx-text-main">
            {language === 'zh' ? 'AI 选股建议' : 'AI Stock Picks'}
          </h1>
          <p className="text-lg text-nofx-text-muted max-w-2xl">
            {language === 'zh'
              ? '我们的 AI 正在分析市场新闻、情绪和技术指标，为您寻找最佳的交易机会。该功能正在开发中，敬请期待！'
              : 'Our AI is analyzing market news, sentiment, and technical indicators to find the best trading opportunities for you. This feature is coming soon!'}
          </p>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mt-12 w-full max-w-4xl">
            <FeatureCard
              icon="📰"
              title={language === 'zh' ? '新闻情绪分析' : 'News Sentiment'}
              desc={language === 'zh' ? '分析来自 NewsAPI 的数千篇文章，判断市场情绪。' : 'Analyzing thousands of articles from NewsAPI to gauge market sentiment.'}
            />
            <FeatureCard
              icon="🧠"
              title={language === 'zh' ? '机器学习预测' : 'ML Prediction'}
              desc={language === 'zh' ? '使用 LSTM 和强化学习模型预测股价趋势。' : 'Using LSTM and Reinforcement Learning models to predict price trends.'}
            />
            <FeatureCard
              icon="🎯"
              title={language === 'zh' ? '智能筛选' : 'Smart Screening'}
              desc={language === 'zh' ? '自动筛选符合您风险偏好和策略的股票。' : 'Automatically screening stocks that match your risk appetite and strategy.'}
            />
          </div>
        </div>
      </Container>
    </DeepVoidBackground>
  )
}

function FeatureCard({ icon, title, desc }: { icon: string, title: string, desc: string }) {
  return (
    <div className="nofx-glass p-6 rounded-xl border border-white/5 hover:border-blue-500/30 transition-all hover:translate-y-[-5px]">
      <div className="text-4xl mb-4">{icon}</div>
      <h3 className="text-xl font-bold text-nofx-text-main mb-2">{title}</h3>
      <p className="text-sm text-nofx-text-muted">{desc}</p>
    </div>
  )
}
