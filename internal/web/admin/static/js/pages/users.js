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
    const confirmMessage = `Are you sure you want to ${actionText} user #${userId}?`;

    if (!confirm(confirmMessage)) {
        return;
    }

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
            alert(`Error: ${data.message || 'Action failed'}`);
        }
    } catch (error) {
        console.error('Error performing user action:', error);
        alert('An unexpected error occurred.');
    }
}
