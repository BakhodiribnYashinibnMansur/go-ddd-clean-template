/**
 * Users Management Scripts
 */
document.addEventListener('DOMContentLoaded', () => {
    initUserActions();
});

function initUserActions() {
    const blockBtns = document.querySelectorAll('.block-user-btn');
    const unblockBtns = document.querySelectorAll('.unblock-user-btn');

    blockBtns.forEach((btn) => {
        btn.addEventListener('click', () => {
            if (btn.dataset.id) {
                handleUserAction(btn.dataset.id, 'block');
            }
        });
    });

    unblockBtns.forEach((btn) => {
        btn.addEventListener('click', () => {
            if (btn.dataset.id) {
                handleUserAction(btn.dataset.id, 'unblock');
            }
        });
    });
}

async function handleUserAction(userId, action) {
    if (!userId) return;

    const actionText = action === 'block' ? 'Block' : 'Unblock';
    const confirmed = await showConfirm({
        title: actionText + ' User',
        message: `Are you sure you want to ${actionText.toLowerCase()} this user?`,
        confirmText: actionText,
        type: action === 'block' ? 'danger' : 'info'
    });

    if (!confirmed) return;

    try {
        const response = await fetch(`/admin/users/${userId}/${action}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            }
        });

        if (response.ok) {
            window.location.reload();
        } else {
            const data = await response.json();
            showToast(data.message || 'Action failed', 'error');
        }
    } catch (error) {
        console.error('Error performing user action:', error);
        showToast('An unexpected error occurred.', 'error');
    }
}
