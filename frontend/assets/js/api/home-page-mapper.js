import { formatRubles, toNumber } from '../ui/formatters.js';

const toAmount = (value) => Number(String(value ?? '').replace(/[^\d]/g, '')) || 0;

const formatMonth = (date = new Date()) => new Intl.DateTimeFormat('ru-RU', { month: 'long' }).format(date);

export const buildHomePageViewModel = ({ dashboard, groups, analytics }) => {
    const monthExpensesValue = toNumber(dashboard.monthExpenses);
    const monthBudget = Math.max(7000, Math.ceil(monthExpensesValue * 1.15 / 100) * 100);
    const budgetProgress = monthBudget ? Math.min(100, Math.round((monthExpensesValue / monthBudget) * 100)) : 0;

    const bars = Array.isArray(analytics?.bars)
        ? analytics.bars.map((bar) => ({ label: bar.label, amount: toAmount(bar.amount ?? bar.value) }))
        : [];

    const previousBar = bars[bars.length - 2]?.amount
        || bars[bars.length - 1]?.amount
        || 0;
    const lastBar = bars[bars.length - 1]?.amount
        || 0;
    const trendPercent = previousBar
        ? Math.round(((lastBar - previousBar) / previousBar) * 100)
        : 0;
    const trendDirection = trendPercent >= 0
        ? 'up' : 'down';

    const categories = Array.isArray(analytics?.categories)
        ? analytics.categories
            .map((item) => ({
                name: item.name,
                amount: toAmount(item.value),
                color: item.color || '#bd5bff'
            }))
            .sort((left, right) => right.amount - left.amount)
        : [];

    const categoriesTotal = categories.reduce((sum, item) => sum + item.amount, 0);
    const topCategories = categories.slice(0, 3).map((category) => ({
        ...category,
        share: categoriesTotal ? Math.round((category.amount / categoriesTotal) * 100) : 0,
        amountLabel: `${formatRubles(category.amount)} ₽`
    }));

    return {
        activeSubscriptions: dashboard.activeSubscriptions,
        expiringSoon: dashboard.expiringSoon,
        familyGroupTitle: dashboard.familyGroup.title,
        familyGroups: groups.length ? groups : [{ name: dashboard.familyGroup.title, members: [] }],
        monthLabel: formatMonth(),
        monthExpensesLabel: `${dashboard.monthExpenses} ₽`,
        monthBudgetLabel: `${formatRubles(monthBudget)} ₽`,
        budgetProgress,
        trendDirection,
        trendPercentLabel: `${trendPercent >= 0
            ? '+'
            : ''}${trendPercent}%`,
        trendLabel: trendPercent >= 0
            ? 'рост к прошлому месяцу'
            : 'снижение к прошлому месяцу',
        topCategories
    };
};
