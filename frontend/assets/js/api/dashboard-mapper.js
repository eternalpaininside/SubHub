import { formatRubles } from '../ui/formatters.js';
import { REMINDER_DAYS } from '../ui/constants.js';

const getMonthlyPrice = (subscription) => {
    const price = Number(subscription.price) || 0;

    return subscription.period === 'год' ? price / 12 : price;
};

const parseDaysLeft = (value) => {
    const match = String(value || '').match(/\d+/);

    return match
        ? Number(match[0])
        : null;
};

const getDaysLeftNumber = (subscription) => {
    const direct = Number(subscription.daysLeftNumber);

    if (Number.isFinite(direct))
        return direct;
    return parseDaysLeft(subscription.daysLeft);
};

export const buildDashboardViewModel = ({ subscriptions }) => {
    const monthExpenses = subscriptions.reduce((sum, subscription) => sum + getMonthlyPrice(subscription), 0);
    const expiringSoon = subscriptions.filter((subscription) => {
        const days = getDaysLeftNumber(subscription);

        return days !== null && days <= REMINDER_DAYS;
    }).length;

    return {
        monthExpenses: formatRubles(monthExpenses),
        activeSubscriptions: subscriptions.length,
        expiringSoon,
        familyGroup: {
            title: 'Мои группы'
        }
    };
};
