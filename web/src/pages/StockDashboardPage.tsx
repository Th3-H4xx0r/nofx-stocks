import { useEffect, useState, useRef } from 'react'
import { mutate } from 'swr'
import { api } from '../lib/api'
import { ChartTabs } from '../components/ChartTabs'
import { DecisionCard } from '../components/DecisionCard'
import { PositionHistory } from '../components/PositionHistory'
import { PunkAvatar, getTraderAvatar } from '../components/PunkAvatar'
import { confirmToast, notify } from '../lib/notify'
import { t, type Language } from '../i18n/translations'
import { LogOut, Loader2 } from 'lucide-react'
import { DeepVoidBackground } from '../components/DeepVoidBackground'
import type {
    SystemStatus,
    AccountInfo,
    Position,
    DecisionRecord,
    Statistics,
    TraderInfo,
    Exchange,
} from '../types'

// --- Helper Functions ---

function getModelDisplayName(modelId: string): string {
    switch (modelId.toLowerCase()) {
        case 'deepseek':
            return 'DeepSeek'
        case 'qwen':
            return 'Qwen'
        case 'claude':
            return 'Claude'
        default:
            return modelId.toUpperCase()
    }
}

// --- Components ---

interface StockDashboardPageProps {
    selectedTrader?: TraderInfo
    traders?: TraderInfo[]
    tradersError?: Error
    selectedTraderId?: string
    onTraderSelect: (traderId: string) => void
    onNavigateToTraders: () => void
    status?: SystemStatus
    account?: AccountInfo
    positions?: Position[]
    decisions?: DecisionRecord[]
    decisionsLimit: number
    onDecisionsLimitChange: (limit: number) => void
    stats?: Statistics
    lastUpdate: string
    language: Language
    exchanges?: Exchange[]
}

export function StockDashboardPage({
    selectedTrader,
    status,
    account,
    positions,
    decisions,
    decisionsLimit,
    onDecisionsLimitChange,
    lastUpdate,
    language,
    traders,
    tradersError,
    selectedTraderId,
    onTraderSelect,
    onNavigateToTraders,
    exchanges,
}: StockDashboardPageProps) {
    const [closingPosition, setClosingPosition] = useState<string | null>(null)
    const [selectedChartSymbol, setSelectedChartSymbol] = useState<string | undefined>(undefined)
    const [chartUpdateKey, setChartUpdateKey] = useState<number>(0)
    const chartSectionRef = useRef<HTMLDivElement>(null)

    // Current positions pagination
    const [positionsPageSize, setPositionsPageSize] = useState<number>(20)
    const [positionsCurrentPage, setPositionsCurrentPage] = useState<number>(1)

    // Calculate paginated positions
    const totalPositions = positions?.length || 0
    const totalPositionPages = Math.ceil(totalPositions / positionsPageSize)
    const paginatedPositions = positions?.slice(
        (positionsCurrentPage - 1) * positionsPageSize,
        positionsCurrentPage * positionsPageSize
    ) || []

    // Reset page when positions change
    useEffect(() => {
        setPositionsCurrentPage(1)
    }, [selectedTraderId, positionsPageSize])

    // Handle symbol click from Decision Card
    const handleSymbolClick = (symbol: string) => {
        // Set the selected symbol
        setSelectedChartSymbol(symbol)
        // Scroll to chart section
        setTimeout(() => {
            chartSectionRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' })
        }, 100)
    }

    // Close Position
    const handleClosePosition = async (symbol: string, side: string) => {
        if (!selectedTraderId) return

        const confirmMsg =
            language === 'zh'
                ? `确定要平仓 ${symbol} ${side === 'LONG' ? '多仓' : '空仓'} 吗？`
                : `Are you sure you want to close ${symbol} ${side === 'LONG' ? 'LONG' : 'SHORT'} position?`

        const confirmed = await confirmToast(confirmMsg, {
            title: language === 'zh' ? '确认平仓' : 'Confirm Close',
            okText: language === 'zh' ? '确认' : 'Confirm',
            cancelText: language === 'zh' ? '取消' : 'Cancel',
        })

        if (!confirmed) return

        setClosingPosition(symbol)
        try {
            await api.closePosition(selectedTraderId, symbol, side)
            notify.success(
                language === 'zh' ? '平仓成功' : 'Position closed successfully'
            )
            // Use SWR mutate to refresh data
            await Promise.all([
                mutate(`positions-${selectedTraderId}`),
                mutate(`account-${selectedTraderId}`),
            ])
        } catch (err: unknown) {
            const errorMsg =
                err instanceof Error
                    ? err.message
                    : language === 'zh'
                        ? '平仓失败'
                        : 'Failed to close position'
            notify.error(errorMsg)
        } finally {
            setClosingPosition(null)
        }
    }

    // If API failed with error
    if (tradersError) {
        return (
            <div className="flex items-center justify-center min-h-[60vh] relative z-10">
                <div className="text-center max-w-md mx-auto px-6">
                    <div className="w-24 h-24 mx-auto mb-6 rounded-full flex items-center justify-center nofx-glass bg-red-500/10 border-red-500/30">
                        <svg className="w-12 h-12 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                        </svg>
                    </div>
                    <h2 className="text-2xl font-bold mb-3 text-nofx-text-main">
                        {language === 'zh' ? '无法连接到服务器' : 'Connection Failed'}
                    </h2>
                    <p className="text-base mb-6 text-nofx-text-muted">
                        {language === 'zh' ? '请确认后端服务已启动。' : 'Please check if the backend service is running.'}
                    </p>
                    <button onClick={() => window.location.reload()} className="px-6 py-3 rounded-lg font-semibold transition-all hover:scale-105 active:scale-95 nofx-glass border border-nofx-gold/30 text-nofx-gold hover:bg-nofx-gold/10">
                        {language === 'zh' ? '重试' : 'Retry'}
                    </button>
                </div>
            </div>
        )
    }

    // If traders list empty
    if (traders && traders.length === 0) {
        return (
            <div className="flex items-center justify-center min-h-[60vh] relative z-10">
                <div className="text-center max-w-md mx-auto px-6">
                    <h2 className="text-2xl font-bold mb-3 text-nofx-text-main">{t('dashboardEmptyTitle', language)}</h2>
                    <p className="text-base mb-6 text-nofx-text-muted">{t('dashboardEmptyDescription', language)}</p>
                    <button onClick={onNavigateToTraders} className="px-6 py-3 rounded-lg font-semibold transition-all hover:scale-105 active:scale-95 nofx-glass border border-nofx-gold/30 text-nofx-gold hover:bg-nofx-gold/10">
                        {t('goToTradersPage', language)}
                    </button>
                </div>
            </div>
        )
    }

    // Loading skeleton
    if (!selectedTrader) {
        return (
            <div className="space-y-6 relative z-10 p-8">
                <div className="nofx-glass p-6 animate-pulse">
                    <div className="h-8 w-48 mb-3 bg-nofx-bg/50 rounded"></div>
                </div>
            </div>
        )
    }

    return (
        <DeepVoidBackground className="min-h-screen pb-12" disableAnimation>
            <div className="w-full px-4 md:px-8 relative z-10 pt-6">
                {/* Trader Header */}
                <div
                    className="mb-6 rounded-lg p-6 animate-scale-in nofx-glass group"
                    style={{
                        background: 'linear-gradient(135deg, rgba(15, 23, 42, 0.6) 0%, rgba(15, 23, 42, 0.4) 100%)',
                    }}
                >
                    <div className="flex items-start justify-between mb-4">
                        <h2 className="text-2xl font-bold flex items-center gap-4 text-nofx-text-main">
                            <div className="relative">
                                <PunkAvatar
                                    seed={getTraderAvatar(selectedTrader.trader_id, selectedTrader.trader_name)}
                                    size={56}
                                    className="rounded-xl border-2 border-blue-500/30 shadow-[0_0_15px_rgba(59,130,246,0.2)]"
                                />
                                <div className="absolute -bottom-1 -right-1 w-4 h-4 bg-nofx-green rounded-full border-2 border-[#0B0E11] shadow-[0_0_8px_rgba(14,203,129,0.8)] animate-pulse" />
                            </div>
                            <div className="flex flex-col">
                                <span className="text-3xl tracking-tight text-nofx-text font-semibold">
                                    {selectedTrader.trader_name}
                                    <span className="ml-3 text-xs bg-blue-500/20 text-blue-400 px-2 py-0.5 rounded border border-blue-500/30 align-middle">US STOCKS</span>
                                </span>
                                <span className="text-xs font-mono text-nofx-text-muted opacity-60 flex items-center gap-2">
                                    <div className="w-1.5 h-1.5 bg-blue-500 rounded-full" />
                                    ID: {selectedTrader.trader_id.slice(0, 8)}...
                                </span>
                            </div>
                        </h2>

                        <div className="flex items-center gap-4">
                            {traders && traders.length > 0 && (
                                <div className="flex items-center gap-2 nofx-glass px-1 py-1 rounded-lg border border-white/5">
                                    <select
                                        value={selectedTraderId}
                                        onChange={(e) => onTraderSelect(e.target.value)}
                                        className="bg-transparent text-sm font-medium cursor-pointer transition-colors text-nofx-text-main focus:outline-none px-2 py-1"
                                    >
                                        {traders.map((trader) => (
                                            <option key={trader.trader_id} value={trader.trader_id} className="bg-[#0B0E11]">
                                                {trader.trader_name}
                                            </option>
                                        ))}
                                    </select>
                                </div>
                            )}
                        </div>
                    </div>
                    <div className="flex items-center gap-6 text-sm flex-wrap text-nofx-text-muted font-mono pl-2">
                        <span className="flex items-center gap-2">
                            <span className="opacity-60">AI Model:</span>
                            <span className="font-bold px-2 py-0.5 rounded text-xs tracking-wide bg-blue-500/10 text-blue-400 border border-blue-500/20">
                                {getModelDisplayName(selectedTrader.ai_model)}
                            </span>
                        </span>
                        <span className="w-px h-3 bg-white/10 hidden md:block" />
                        <span className="flex items-center gap-2">
                            <span className="opacity-60">Exchange:</span>
                            <span className="text-nofx-text-main font-semibold">Alpaca</span>
                        </span>
                        {status && (
                            <div className="hidden md:contents">
                                <span className="w-px h-3 bg-white/10" />
                                <span>Cycles: <span className="text-nofx-text-main">{status.call_count}</span></span>
                            </div>
                        )}
                    </div>
                </div>

                {/* Account Overview */}
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
                    <StatCard
                        title={t('totalEquity', language)}
                        value={`$${account?.total_equity?.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) || '0.00'}`}
                        unit="USD"
                        change={account?.total_pnl_pct || 0}
                        positive={(account?.total_pnl ?? 0) > 0}
                        icon="💰"
                    />
                    <StatCard
                        title={t('availableBalance', language)}
                        value={`$${account?.available_balance?.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) || '0.00'}`}
                        unit="USD"
                        subtitle={`${account?.available_balance && account?.total_equity ? ((account.available_balance / account.total_equity) * 100).toFixed(1) : '0.0'}% Cash`}
                        icon="💳"
                    />
                    <StatCard
                        title={t('totalPnL', language)}
                        value={`${account?.total_pnl !== undefined && account.total_pnl >= 0 ? '+' : ''}$${account?.total_pnl?.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) || '0.00'}`}
                        unit="USD"
                        change={account?.total_pnl_pct || 0}
                        positive={(account?.total_pnl ?? 0) >= 0}
                        icon="📈"
                    />
                    <StatCard
                        title="Day Trades"
                        value={`${account?.day_trade_count || 0}`}
                        unit="/ 3 (PDT)"
                        subtitle="Pattern Day Trader Count"
                        icon="📅"
                    />
                </div>

                {/* Main Content Area */}
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
                    {/* Left Column: Charts + Positions */}
                    <div className="space-y-6">
                        {/* Chart Tabs */}
                        <div
                            ref={chartSectionRef}
                            className="chart-container animate-slide-in scroll-mt-32 backdrop-blur-sm"
                            style={{ animationDelay: '0.1s' }}
                        >
                            <ChartTabs
                                traderId={selectedTrader.trader_id}
                                selectedSymbol={selectedChartSymbol}
                                updateKey={chartUpdateKey}
                                exchangeId="alpaca" // Force Alpaca charts
                            />
                        </div>

                        {/* Current Positions */}
                        <div className="nofx-glass p-6 animate-slide-in relative overflow-hidden group" style={{ animationDelay: '0.15s' }}>
                            <div className="absolute top-0 right-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                                <div className="w-24 h-24 rounded-full bg-blue-500 blur-3xl" />
                            </div>
                            <div className="flex items-center justify-between mb-5 relative z-10">
                                <h2 className="text-lg font-bold flex items-center gap-2 text-nofx-text-main uppercase tracking-wide">
                                    <span className="text-blue-500">◈</span> {t('currentPositions', language)}
                                </h2>
                                {positions && positions.length > 0 && (
                                    <div className="text-xs px-2 py-1 rounded bg-blue-500/10 text-blue-400 border border-blue-500/20 font-mono">
                                        {positions.length} {t('active', language)}
                                    </div>
                                )}
                            </div>
                            {positions && positions.length > 0 ? (
                                <div>
                                    <div className="overflow-x-auto">
                                        <table className="w-full text-xs">
                                            <thead className="text-left border-b border-white/5">
                                                <tr>
                                                    <th className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-left">{t('symbol', language)}</th>
                                                    <th className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-center">{t('side', language)}</th>
                                                    <th className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-center">{language === 'zh' ? '操作' : 'Action'}</th>
                                                    <th className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-right hidden md:table-cell">{language === 'zh' ? '入场价' : 'Entry'}</th>
                                                    <th className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-right hidden md:table-cell">{language === 'zh' ? '现价' : 'Price'}</th>
                                                    <th className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-right">{language === 'zh' ? '数量' : 'Qty'}</th>
                                                    <th className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-right hidden md:table-cell">{language === 'zh' ? '市值' : 'Mkt Val'}</th>
                                                    <th className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-right">{language === 'zh' ? '盈亏' : 'P&L'}</th>
                                                </tr>
                                            </thead>
                                            <tbody>
                                                {paginatedPositions.map((pos, i) => (
                                                    <tr
                                                        key={i}
                                                        className="border-b border-white/5 last:border-0 transition-all hover:bg-white/5 cursor-pointer group/row"
                                                        onClick={() => {
                                                            setSelectedChartSymbol(pos.symbol)
                                                            setChartUpdateKey(Date.now())
                                                            if (chartSectionRef.current) {
                                                                chartSectionRef.current.scrollIntoView({ behavior: 'smooth', block: 'start' })
                                                            }
                                                        }}
                                                    >
                                                        <td className="px-1 py-3 font-mono font-semibold whitespace-nowrap text-left text-nofx-text-main group-hover/row:text-white transition-colors">
                                                            {pos.symbol}
                                                        </td>
                                                        <td className="px-1 py-3 whitespace-nowrap text-center">
                                                            <span className={`px-1.5 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider ${pos.side === 'long' ? 'bg-nofx-green/10 text-nofx-green' : 'bg-nofx-red/10 text-nofx-red'}`}>
                                                                {pos.side}
                                                            </span>
                                                        </td>
                                                        <td className="px-1 py-3 whitespace-nowrap text-center">
                                                            <button
                                                                type="button"
                                                                onClick={(e) => {
                                                                    e.stopPropagation()
                                                                    handleClosePosition(pos.symbol, pos.side.toUpperCase())
                                                                }}
                                                                disabled={closingPosition === pos.symbol}
                                                                className="inline-flex items-center gap-1 px-2 py-1 rounded text-[10px] font-semibold transition-all hover:scale-105 disabled:opacity-50 disabled:cursor-not-allowed mx-auto bg-nofx-red/10 text-nofx-red border border-nofx-red/30 hover:bg-nofx-red/20"
                                                            >
                                                                {closingPosition === pos.symbol ? (
                                                                    <Loader2 className="w-3 h-3 animate-spin" />
                                                                ) : (
                                                                    <LogOut className="w-3 h-3" />
                                                                )}
                                                                {language === 'zh' ? '平仓' : 'Close'}
                                                            </button>
                                                        </td>
                                                        <td className="px-1 py-3 font-mono whitespace-nowrap text-right text-nofx-text-main hidden md:table-cell">${pos.entry_price.toFixed(2)}</td>
                                                        <td className="px-1 py-3 font-mono whitespace-nowrap text-right text-nofx-text-main hidden md:table-cell">${pos.mark_price.toFixed(2)}</td>
                                                        <td className="px-1 py-3 font-mono whitespace-nowrap text-right text-nofx-text-main">{pos.quantity}</td>
                                                        <td className="px-1 py-3 font-mono font-bold whitespace-nowrap text-right text-nofx-text-main hidden md:table-cell">${(pos.quantity * pos.mark_price).toFixed(2)}</td>
                                                        <td className="px-1 py-3 font-mono whitespace-nowrap text-right">
                                                            <span className={`font-bold ${pos.unrealized_pnl >= 0 ? 'text-nofx-green' : 'text-nofx-red'}`}>
                                                                {pos.unrealized_pnl >= 0 ? '+' : ''}{pos.unrealized_pnl.toFixed(2)}
                                                            </span>
                                                        </td>
                                                    </tr>
                                                ))}
                                            </tbody>
                                        </table>
                                    </div>
                                    {/* Pagination (Simplified) */}
                                    {totalPositions > 20 && (
                                        <div className="flex justify-between mt-4 text-xs text-nofx-text-muted">
                                            <span>Showing {paginatedPositions.length} of {totalPositions}</span>
                                            <div className="flex gap-2">
                                                <button onClick={() => setPositionsCurrentPage(p => Math.max(1, p - 1))} disabled={positionsCurrentPage === 1}>Prev</button>
                                                <button onClick={() => setPositionsCurrentPage(p => Math.min(totalPositionPages, p + 1))} disabled={positionsCurrentPage === totalPositionPages}>Next</button>
                                            </div>
                                        </div>
                                    )}
                                </div>
                            ) : (
                                <div className="text-center py-16 text-nofx-text-muted opacity-60">
                                    <div className="text-6xl mb-4 opacity-50 grayscale">📊</div>
                                    <div className="text-lg font-semibold mb-2">{t('noPositions', language)}</div>
                                    <div className="text-sm">{t('noActivePositions', language)}</div>
                                </div>
                            )}
                        </div>
                    </div>

                    {/* Right Column: Recent Decisions */}
                    <div className="nofx-glass p-6 animate-slide-in h-fit lg:sticky lg:top-24 flex flex-col" style={{ animationDelay: '0.2s' }}>
                        {/* Header */}
                        <div className="flex items-center gap-3 mb-5 pb-4 border-b border-white/5 shrink-0">
                            <div className="w-10 h-10 rounded-xl flex items-center justify-center text-xl shadow-[0_4px_14px_rgba(99,102,241,0.4)]" style={{ background: 'linear-gradient(135deg, #6366F1 0%, #8B5CF6 100%)' }}>
                                🧠
                            </div>
                            <div className="flex-1">
                                <h2 className="text-xl font-bold text-nofx-text-main">{t('recentDecisions', language)}</h2>
                            </div>
                            <select
                                value={decisionsLimit}
                                onChange={(e) => onDecisionsLimitChange(Number(e.target.value))}
                                className="px-3 py-1.5 rounded-lg text-sm bg-black/40 text-nofx-text-main border border-white/10"
                            >
                                <option value={5}>5</option>
                                <option value={10}>10</option>
                                <option value={20}>20</option>
                            </select>
                        </div>

                        {/* Decisions List */}
                        <div className="space-y-4 overflow-y-auto pr-2 custom-scrollbar" style={{ maxHeight: 'calc(100vh - 280px)' }}>
                            {decisions && decisions.length > 0 ? (
                                decisions.map((decision, i) => (
                                    <DecisionCard key={i} decision={decision} language={language} onSymbolClick={handleSymbolClick} />
                                ))
                            ) : (
                                <div className="py-16 text-center text-nofx-text-muted opacity-60">
                                    <div className="text-lg font-semibold mb-2 text-nofx-text-main">{t('noDecisionsYet', language)}</div>
                                    <div className="text-sm">{t('aiDecisionsWillAppear', language)}</div>
                                </div>
                            )}
                        </div>
                    </div>
                </div>

                {/* Position History */}
                {selectedTraderId && (
                    <div className="nofx-glass p-6 animate-slide-in" style={{ animationDelay: '0.25s' }}>
                        <div className="flex items-center justify-between mb-5">
                            <h2 className="text-xl font-bold flex items-center gap-2 text-nofx-text-main">
                                <span className="text-2xl">📜</span>
                                {t('positionHistory.title', language)}
                            </h2>
                        </div>
                        <PositionHistory traderId={selectedTraderId} />
                    </div>
                )}
            </div>
        </DeepVoidBackground>
    )
}

function StatCard({ title, value, unit, change, positive, subtitle, icon }: any) {
    return (
        <div className="group nofx-glass p-5 rounded-lg transition-all duration-300 hover:bg-white/5 hover:translate-y-[-2px] border border-white/5 hover:border-blue-500/20 relative overflow-hidden">
            <div className="absolute top-0 right-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity text-4xl grayscale group-hover:grayscale-0">{icon}</div>
            <div className="text-xs mb-2 font-mono uppercase tracking-wider text-nofx-text-muted flex items-center gap-2">{title}</div>
            <div className="flex items-baseline gap-1 mb-1">
                <div className="text-2xl font-bold font-mono text-nofx-text-main tracking-tight group-hover:text-white transition-colors">{value}</div>
                {unit && <span className="text-xs font-mono text-nofx-text-muted opacity-60">{unit}</span>}
            </div>
            {change !== undefined && (
                <div className="flex items-center gap-1">
                    <div className={`text-sm mono font-bold flex items-center gap-1 ${positive ? 'text-nofx-green' : 'text-nofx-red'}`}>
                        <span>{positive ? '▲' : '▼'}</span>
                        <span>{positive ? '+' : ''}{change.toFixed(2)}%</span>
                    </div>
                </div>
            )}
            {subtitle && <div className="text-xs mt-2 mono text-nofx-text-muted opacity-80">{subtitle}</div>}
        </div>
    )
}
