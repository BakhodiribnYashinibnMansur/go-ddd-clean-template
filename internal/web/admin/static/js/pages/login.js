/**
 * Login Page Scripts
 */
document.addEventListener('DOMContentLoaded', () => {
    // Password Toggle
    const togglePassword = document.querySelector('.toggle-password');
    const passwordInput = document.querySelector('input[type="password"]');

    if (togglePassword && passwordInput) {
        togglePassword.addEventListener('click', function () {
            const type = passwordInput.getAttribute('type') === 'password' ? 'text' : 'password';
            passwordInput.setAttribute('type', type);

            // Toggle icon class
            if (type === 'password') {
                this.classList.remove('bx-show');
                this.classList.add('bx-hide');
            } else {
                this.classList.remove('bx-hide');
                this.classList.add('bx-show');
            }
        });
    }

    // Loader on submit
    const form = document.querySelector('form');
    const submitBtn = document.querySelector('.btn-primary');

    if (form && submitBtn) {
        form.addEventListener('submit', () => {
            submitBtn.classList.add('loading');
        });
    }
});
