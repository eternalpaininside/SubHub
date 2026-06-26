import { API_BASE_URL, REQUEST_TIMEOUT_MS } from './runtime-config.js';
import { buildAnalyticsPageViewModel } from './analytics-page-mapper.js';
import { buildDashboardViewModel } from './dashboard-mapper.js';
import { buildHomePageViewModel } from './home-page-mapper.js';
import { getCurrentUser, getCurrentUserId } from '../ui/session.js';
import { formatRubles } from '../ui/formatters.js';

const analyticsPalette = ['#bd5bff', '#ff5da5', '#42a3ff', '#ffb128', '#39d17f', '#7cd4ff'];
const categoryColorMap = {
    'Стриминг':       'linear-gradient(135deg, #240000, #7d0014)',
    'Музыка':         'linear-gradient(135deg, #0b3a28, #0f8d50)',
    'Облако':         'linear-gradient(135deg, #00132f, #004ab8)',
    'Продуктивность': 'linear-gradient(135deg, #1e1e22, #2a2a33)',
    'AI':             'linear-gradient(135deg, #0c3f34, #18a67f)',
    'Комплекс':       'linear-gradient(135deg, #421300, #8d2314)'
};

async function request(path, options = {}) {
    const controller = new AbortController();
    const timeoutId = window.setTimeout(() =>
        controller.abort(), REQUEST_TIMEOUT_MS);

    try {
        const response = await fetch(`${API_BASE_URL}${path}`, {
            headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
            ...options,
            signal: controller.signal
        });

        let payload = null;
        try {
            payload = await response.json();
        } catch {
            payload = null;
        }

        if (!response.ok)
            throw new Error(payload?.error || `HTTP ${response.status}`);
        return payload;
    } finally {
        window.clearTimeout(timeoutId);
    }
}

const buildQuery = (params = {}) => {
    const query = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '')
            query.set(key, String(value));
    });
    const suffix = query.toString();
    return suffix ? `?${suffix}` : '';
};

const requireUserId = () => {
    const userId = getCurrentUserId();
    if (!userId)
        throw new Error('Требуется вход в систему');
    return userId;
};

const requestForCurrentUser = (path) => request(`${path}${buildQuery({ user_id: requireUserId() })}`);
const normalizePeriod = (value) => Number(value) === 12 ? 'год' : 'мес';
const pickIcon = (name) => String(name || '?').trim().charAt(0).toUpperCase() || '?';

const formatDateText = (isoDate) => {
    if (!isoDate)
        return '—';
    const date = new Date(`${isoDate}T00:00:00`);
    if (Number.isNaN(date.getTime()))
        return isoDate;
    return new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' }).format(date);
};

const getDaysLeftNumber = (isoDate) => {
    if (!isoDate)
        return null;
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const target = new Date(`${isoDate}T00:00:00`);
    if (Number.isNaN(target.getTime()))
        return null;
    return Math.ceil((target.getTime() - today.getTime()) / 86400000);
};

const getMonthlyPrice = ({ price, period }) => {
    const amount = Number(price) || 0;
    return period === 'год' ? Math.round(amount / 12) : amount;
};

const buildSubscriptionsPageViewModel = (subscriptions) => {
    const monthlyTotal = subscriptions.reduce((sum, subscription) => sum + getMonthlyPrice(subscription), 0);

    return {
        items: subscriptions,
        subtitle: `${subscriptions.length} активных · ${formatRubles(monthlyTotal)} ₽/мес`
    };
};

const mapCategories = (analytics) => (analytics.by_category || []).map((item, index) => ({
    name: item.category,
    value: `${item.total} ₽`,
    color: analyticsPalette[index % analyticsPalette.length]
}));

const mapSubscription = (subscription) => {
    const daysLeftNumber = getDaysLeftNumber(subscription.next_payment);
    return {
        id: subscription.id,
        name: subscription.name,
        icon: pickIcon(subscription.name),
        iconBg: categoryColorMap[subscription.category] ||
            'linear-gradient(135deg, #25324b, #3d5a80)',
        price: String(subscription.price),
        period: normalizePeriod(subscription.period),
        category: subscription.category,
        planType: subscription.plan_type || 'Индивидуальный',
        dateText: formatDateText(subscription.next_payment),
        nextPaymentDate: subscription.next_payment, daysLeftNumber,
        daysLeft: daysLeftNumber === null ? '' : `${daysLeftNumber} дн.`,
        siteUrl: subscription.link,
        comment: subscription.comment || '',
        status: Boolean(subscription.status)
    };
};

const mapGroup = (group) => ({
    ...group,
    price: String(group.price || 0),
    period: normalizePeriod(group.period),
    inviteUrl: group.invite_url || '',
    members: Array.isArray(group.members) ? group.members : [],
    services: Array.isArray(group.services) ? group.services : [],
    subscriptionIds: Array.isArray(group.subscription_ids) ?
        group.subscription_ids.map((id) => Number(id)).filter(Boolean) : []
});

const mapProfile = (profile) => {
    const currentUser = ае();
    return {
        user: {
            name: profile.user.name,
            handle: currentUser?.email ?
                `@${String(currentUser.email).split('@')[0]}` : `@user${profile.user.id}`,
            email: profile.user.email
        },
        stats: [
            { label: 'Активных подписок', value: String(profile.stats.active_subscriptions) },
            { label: 'Групп', value: String(profile.stats.group_count) },
            { label: 'Расходы в месяц', value: `${profile.stats.monthly_spend} ₽` }
        ]
    };
};

const getSubscriptions = async (filters = {}) => {
    const subscriptions = (await requestForCurrentUser('/subscriptions')).map(mapSubscription);

    return subscriptions.filter((subscription) => {
        const matchesCategory = !filters.category
            || filters.category === 'Все' || subscription.category === filters.category;
        const matchesPlanType = !filters.planType
            || subscription.planType === filters.planType;
        return matchesCategory && matchesPlanType;
    });
};

const getAnalytics = async () => requestForCurrentUser('/analytics');
const getGroups = async () => (await requestForCurrentUser('/groups')).map(mapGroup);
const getProfile = async () => mapProfile(await requestForCurrentUser('/profile'));

export const api = {
    register: (payload) => request('/auth/register',
        { method: 'POST', body: JSON.stringify(payload) }),
    login: (payload) => request('/auth/login',
        { method: 'POST', body: JSON.stringify(payload) }),

    getSubscriptions,

    getSubscriptionsPage: async (filters = {}) => buildSubscriptionsPageViewModel(await getSubscriptions(filters)),

    createSubscription: (payload) => request('/subscriptions', {
        method: 'POST',
        body: JSON.stringify({ ...payload, user_id: payload.user_id || requireUserId() })
    }),

    updateSubscription: (id, payload) => request(`/subscriptions/${id}`, {
        method: 'PUT',
        body: JSON.stringify({ ...payload, user_id: payload.user_id || requireUserId() })
    }),

    deleteSubscription: (id) => request(`/subscriptions/${id}${
        buildQuery({user_id: requireUserId()})}`, { method: 'DELETE' }),

    getAnalytics,

    getAnalyticsPage: async () => {
        const [analytics, subscriptions, profile, groups] = await Promise.all([
            getAnalytics(),
            getSubscriptions(),
            getProfile(),
            getGroups().catch(() => [])
        ]);

        return buildAnalyticsPageViewModel({
            analytics: {
                month: String(analytics.monthly_total),
                bars: Array.isArray(analytics.bars) ? analytics.bars : [],
                categories: mapCategories(analytics)
            },
            subscriptions,
            profile,
            groups: groups.map((group) => ({ ...group, monthlyPrice: getMonthlyPrice(group) }))
        });
    },

    getGroups,

    createGroup: (payload) => request('/groups', {
        method: 'POST',
        body: JSON.stringify({ ...payload, owner_id: payload.owner_id || requireUserId() })
    }),

    updateGroup: (id, payload) => request(`/groups/${id}`, {
        method: 'PUT',
        body: JSON.stringify({ ...payload, owner_id: payload.owner_id || requireUserId() })
    }),

    deleteGroup: (id) => request(`/groups/${id}${buildQuery({ owner_id: requireUserId() })}`, {
        method: 'DELETE'
    }),

    joinGroup: (inviteURL) => request('/groups/join', {
        method: 'POST',
        body: JSON.stringify({ user_id: requireUserId(), invite_url: inviteURL })
    }),

    getProfile,

    getHomePage: async () => {
        const [subscriptions, analytics, groups] = await Promise.all([
            getSubscriptions(),
            getAnalytics(),
            getGroups().catch(() => [])
        ]);

        return buildHomePageViewModel({
            dashboard: buildDashboardViewModel({ subscriptions}),
            groups,
            analytics: {
                bars: Array.isArray(analytics.bars) ? analytics.bars : [],
                categories: mapCategories(analytics)
            }
        });
    }
};
