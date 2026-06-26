import {formatRubles, toNumber} from '../ui/formatters.js';

const getMonthlyPrice = (subscription) => {
    const price = Number(subscription.price) || 0;
    
    return subscription.period === 'год' ? Math.round(price / 12) : price;
};

const parseDaysLeft = (value) => {
    const match = String(value || '').match(/\d+/);
    
    return match ? Number(match[0]) : null;
};

const getDaysLeftNumber = (subscription) => {
    const direct = Number(subscription.daysLeftNumber);

    if (Number.isFinite(direct))
        return direct;

    return parseDaysLeft(subscription.daysLeft);
};

const getNiceStep = (value) => {
    const safeValue = Math.max(1, value);
    const exponent = Math.floor(Math.log10(safeValue));
    const base = 10 ** exponent;
    const fraction = safeValue / base;

    if (fraction <= 1)
        return base;
    if (fraction <= 2)
        return 2 * base;
    if (fraction <= 5)
        return 5 * base;
    return 10 * base;
};

const polarToCartesian = (cx, cy, radius, angleDeg) => {
    const angleRad = (angleDeg * Math.PI) / 180;
   
    return {
        x: cx + radius * Math.cos(angleRad),
        y: cy + radius * Math.sin(angleRad)
    };
};

const describePieSlice = (cx, cy, radius, startAngle, endAngle) => {
    const start = polarToCartesian(cx, cy, radius, startAngle);
    const end = polarToCartesian(cx, cy, radius, endAngle);
    const largeArcFlag = endAngle - startAngle > 180 ? 1 : 0;

    return `M ${cx} ${cy} L ${start.x} ${start.y} A ${radius} ${radius} 0 ${largeArcFlag} 1 ${end.x} ${end.y} Z`;
};

const formatSubscriptionsPreview = (names) => {
    if (!names.length)
        return 'Нет подписок';

    const firstThree = names.slice(0, 3);
    const rest = names.length - firstThree.length;
    return rest > 0 ? `${firstThree.join(', ')}, ...+${rest}` : firstThree.join(', ');
};

export const buildAnalyticsPageViewModel = ({ analytics, subscriptions, profile, groups = [] }) => {
    const bars = analytics.bars.map((bar) => ({
        label: bar.label,
        amount: toNumber(bar.amount ?? bar.value)
    }));

    const categoryTotals = analytics.categories
        .map((category) => ({ ...category, amount: toNumber(category.value) }))
        .sort((left, right) => right.amount - left.amount);
    const totalCategoriesAmount = categoryTotals.reduce((sum, item) => sum + item.amount, 0);
    const topCategory = categoryTotals[0] ||
        { name: '—', amount: 0, color: '#bd5bff' };
    const topCategoryShare = totalCategoriesAmount
        ? Math.round((topCategory.amount / totalCategoriesAmount) * 100)
        : 0;

    const categorySubscriptions = subscriptions.reduce((accumulator, subscription) => {
        const key = String(subscription.category || '').trim().toLowerCase();
        if (!key)
            return accumulator;
        if (!accumulator.has(key))
            accumulator.set(key, []);

        accumulator.get(key).push(subscription.name);
        return accumulator;
    }, new Map());

    const ringCategories = analytics.categories
        .map((category) => ({ ...category, amount: toNumber(category.value) }))
        .filter((category) => category.amount > 0);
    const ringTotal = ringCategories.reduce((sum, item) => sum + item.amount, 0);
    const pieRadius = 52;
   
    let pieCursor = -90;
    const ringSegments = ringCategories.map((item) => {
        const shareRaw = ringTotal ? (item.amount / ringTotal) * 100 : 0;
        const sweep = (shareRaw / 100) * 360;
        const startAngle = pieCursor;
        const endAngle = pieCursor + sweep;
        const names = categorySubscriptions.get(String(item.name).trim().toLowerCase()) || [];

        const segment = {
            name: item.name,
            color: item.color,
            share: Math.round(shareRaw),
            amount: item.amount,
            amountLabel: `${formatRubles(item.amount)} ₽`,
            path: describePieSlice(60, 60, pieRadius, startAngle, endAngle),
            subscriptionsPreview: formatSubscriptionsPreview(names)
        };
        pieCursor = endAngle;
        
        return segment;
    });

    const subscriptionStats = subscriptions
        .map((subscription) => ({
            ...subscription,
            monthlyPrice: getMonthlyPrice(subscription),
            daysLeftNumber: getDaysLeftNumber(subscription)
        }))
        .sort((left, right) => right.monthlyPrice - left.monthlyPrice);

    const mostExpensiveSubscription = subscriptionStats[0];
    const upcomingCharges = subscriptionStats
        .filter((subscription) => subscription.daysLeftNumber !== null && subscription.daysLeftNumber <= 7)
        .sort((left, right) => left.daysLeftNumber - right.daysLeftNumber)
        .slice(0, 4);

    const groupedTotal = subscriptionStats
        .filter((subscription) => String(subscription.planType).toLowerCase().includes('груп'))
        .reduce((sum, subscription) => sum + subscription.monthlyPrice, 0);

    const individualTotal = subscriptionStats
        .filter((subscription) => !String(subscription.planType).toLowerCase().includes('груп'))
        .reduce((sum, subscription) => sum + subscription.monthlyPrice, 0);

    const yearlyProjection = (groupedTotal + individualTotal) * 12;
    const currentMonthAmount = toNumber(analytics.month);
    const previousBar = bars[bars.length - 2]?.amount
        ?? bars[bars.length - 1]?.amount
        ?? 0;

    const dynamicTrend = Number((previousBar
        ? ((currentMonthAmount - previousBar) / previousBar) * 100
        : 0).toFixed(1));

    const spentYearToDate = bars.reduce((sum, bar) =>
        sum + bar.amount, 0);
    const monthsWithHistory = bars.filter((bar) =>
        bar.amount > 0).length || bars.length || 1;

    const averagePerMonth = Math.round(spentYearToDate / monthsWithHistory);
    const upcomingTotal = upcomingCharges.reduce((sum, subscription) =>
        sum + subscription.monthlyPrice, 0);

    const nearestChargeDays = upcomingCharges[0]?.daysLeftNumber ?? null;

    const maxAmount = Math.max(...bars.map((bar) => bar.amount), 1);
    const userMaxBudget = toNumber(profile?.plan?.monthlyBudget);
    const yTickSegments = userMaxBudget > 0
        ? 5 : 4;

    const autoStep = getNiceStep(maxAmount / yTickSegments);
    const autoMax = autoStep * yTickSegments;

    const yMax = userMaxBudget > 0
        ? Math.max(maxAmount, userMaxBudget)
        : autoMax;
    const yStep = userMaxBudget > 0
        ? yMax / yTickSegments
        : autoStep;

    const yTicks = Array.from({ length: yTickSegments + 1 },
        (_, index) => {
            const value = yMax - index * yStep;

            return { label: `${formatRubles(value)} ₽` };
        });

    return {
        topCards: {
            month: {
                valueLabel: `${analytics.month} ₽`,
                trendClass: dynamicTrend > 0 ? 'is-up' : dynamicTrend < 0 ? 'is-down' : 'is-neutral',
                trendText: dynamicTrend === 0
                    ? 'Как в прошлом месяце'
                    : `На ${Math.abs(dynamicTrend)}% ${dynamicTrend > 0 ? 'больше' : 'меньше'} прошлого месяца`,
                previousSpendText: `Траты в прошлом месяце: ${formatRubles(previousBar)} ₽`
            },
            year: {
                valueLabel: `${formatRubles(spentYearToDate)} ₽`,
                averageText: `Среднее в месяц: ${formatRubles(averagePerMonth)} ₽`,
                projectionText: `Может быть потрачено за год: ${formatRubles(yearlyProjection)} ₽`
            },
            week: {
                title: 'Списания на неделе',
                valueLabel: `${formatRubles(upcomingTotal)} ₽`,
                primaryText: upcomingCharges.length
                    ? `Ближайшее списание через ${nearestChargeDays} дн.`
                    : 'На этой неделе списаний нет',
                secondaryText: upcomingCharges.length
                    ? `Запланировано списаний: ${upcomingCharges.length}`
                    : 'Запланированных списаний: 0'
            }
        },
        expensesChart: {
            yTicks,
            bars: bars.map((bar, index) => ({
                label: bar.label,
                amount: bar.amount,
                amountLabel: `${formatRubles(bar.amount)} ₽`,
                heightPercent: Math.max(0, (bar.amount / yMax) * 100),
                delayMs: index * 85
            }))
        },
        categoriesPie: {
            segments: ringSegments,
            legend: ringCategories.map((item) => {
                const share = ringTotal ? Math.round((item.amount / ringTotal) * 100) : 0;
                return {
                    name: item.name,
                    color: item.color,
                    amountLabel: `${formatRubles(item.amount)} ₽`,
                    share,
                    meterWidth: Math.max(share, 4)
                };
            })
        },
        facts: {
            topCategory: {
                name: topCategory.name,
                amountLabel: `${formatRubles(topCategory.amount)} ₽`,
                share: topCategoryShare,
                color: topCategory.color,
                progressWidth: Math.max(topCategoryShare, 8)
            },
            mostExpensiveSubscription: {
                name: mostExpensiveSubscription?.name || '—',
                priceLabel: `${formatRubles(mostExpensiveSubscription?.monthlyPrice || 0)} ₽/мес`,
                planType: mostExpensiveSubscription?.planType || '—',
                category: mostExpensiveSubscription?.category || '—'
            },
            upcomingCharges: upcomingCharges.map((subscription) => ({
                name: subscription.name,
                daysLeftNumber: subscription.daysLeftNumber
            })),
            planSplit: {
                familyLabel: `${formatRubles(groupedTotal)} ₽`,
                individualLabel: `${formatRubles(individualTotal)} ₽`
            }
        }
    };
};
