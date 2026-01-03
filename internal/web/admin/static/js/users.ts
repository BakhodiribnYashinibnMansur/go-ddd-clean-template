/**
 * Users Management Scripts
 */
document.addEventListener('DOMContentLoaded', () => {
    initUserActions();
});

function initUserActions(): void {
    const blockBtns = document.querySelectorAll('.block-user-btn');
    const unblockBtns = document.querySelectorAll('.unblock-user-btn');

    blockBtns.forEach((btn) => {
        const element = btn as HTMLElement;
        element.addEventListener('click', () => {
            if (element.dataset.id) {
                handleUserAction(element.dataset.id, 'block');
            }
        });
    });

    unblockBtns.forEach((btn) => {
        const element = btn as HTMLElement;
        element.addEventListener('click', () => {
            if (element.dataset.id) {
                handleUserAction(element.dataset.id, 'unblock');
            }
        });
    });
}

async function handleUserAction(userId: string, action: 'block' | 'unblock'): Promise<void> {
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
                // 'X-CSRF-Token': getCsrfToken() // function to get token from meta or cookie
            }
        });

        if (response.ok) {
            window.location.reload();
        } else {
            const data: { message?: string } = await response.json();
            alert(`Error: ${data.message || 'Action failed'}`);
        }
    } catch (error) {
        console.error('Error performing user action:', error);
        alert('An unexpected error occurred.');
    }
}
