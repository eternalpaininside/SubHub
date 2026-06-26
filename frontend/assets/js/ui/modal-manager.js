export const createModalManager = ({ modalSelector = '.modal-backdrop' } = {}) => {
    const modals = new Map();

    document.querySelectorAll(modalSelector).forEach((modal) => {
        const id = modal.id;
        if (!id) return;
        modal.style.display = 'none';
        modals.set(id, modal);
    });

    const open = (id) => {
        const modal = modals.get(id);
        if (!modal) return;
        modal.style.display = 'flex';
        requestAnimationFrame(() => {
            modal.classList.add('is-open');
            modal.setAttribute('aria-hidden', 'false');
            document.body.classList.add('modal-open');
        });
    };

    const close = (idOrModal) => {
        const modal = typeof idOrModal === 'string' ? modals.get(idOrModal) : idOrModal;
        if (!modal) return;
        if (modal.dataset.locked === 'true') return;

        modal.classList.remove('is-open');
        modal.setAttribute('aria-hidden', 'true');

        window.setTimeout(() => {
            if (!modal.classList.contains('is-open')) {
                modal.style.display = 'none';
            }
            if (!document.querySelector('.modal-backdrop.is-open')) {
                document.body.classList.remove('modal-open');
            }
        }, 280);
    };

    const closeAll = () => {
        modals.forEach((modal) => {
            if (modal.classList.contains('is-open')) close(modal);
        });
    };

    const bindTrigger = ({ selector, modalId, beforeOpen }) => {
        document.addEventListener('click', (event) => {
            const trigger = event.target instanceof Element ? event.target.closest(selector) : null;
            if (!trigger) return;
            if (typeof beforeOpen === 'function') {
                beforeOpen(trigger);
            }
            open(modalId);
        });
    };

    const bindCloseHandlers = () => {
        document.querySelectorAll('.js-close-modal').forEach((btn) => {
            btn.addEventListener('click', () => close(btn.closest('.modal-backdrop')));
        });

        document.querySelectorAll('.modal-backdrop').forEach((backdrop) => {
            backdrop.addEventListener('click', (event) => {
                if (event.target === backdrop) close(backdrop);
            });
        });

        document.addEventListener('keydown', (event) => {
            if (event.key === 'Escape') closeAll();
        });
    };

    return {
        open,
        close,
        closeAll,
        getModal: (id) => modals.get(id),
        bindTrigger,
        bindCloseHandlers
    };
};
