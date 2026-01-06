document.addEventListener('DOMContentLoaded', () => {
    const loginForm = document.getElementById('loginForm');
    const passwordInput = document.getElementById('password');
    const togglePasswordBtn = document.getElementById('togglePassword');
    const loginBtn = document.getElementById('loginBtn');

    // Toggle Password Visibility
    if (togglePasswordBtn && passwordInput) {
        togglePasswordBtn.addEventListener('click', () => {
            const type = passwordInput.getAttribute('type') === 'password' ? 'text' : 'password';
            passwordInput.setAttribute('type', type);

            // Toggle Icon
            togglePasswordBtn.classList.toggle('bx-hide');
            togglePasswordBtn.classList.toggle('bx-show');
        });
    }

    // Handle Form Submission (Animation)
    if (loginForm) {
        loginForm.addEventListener('submit', (e) => {
            // Check validity
            if (!loginForm.checkValidity()) {
                e.preventDefault();
                // Find first invalid input and focus
                const invalidInput = loginForm.querySelector(':invalid');
                if (invalidInput) invalidInput.focus();
                return;
            }

            // Show Loading State
            loginBtn.classList.add('loading');
            loginBtn.setAttribute('disabled', 'true');

            // Allow form to submit normally after a slight delay for effect? 
            // Or just let it go. If we want to see the animation, we might need a tiny delay or just let the browser handle it.
            // Since it's a synchronous submit, the page will unload. The user will see the spinner briefly.
            // No e.preventDefault() here means it submits.
        });
    }

    // Input animations helper (if needed)
    const inputs = document.querySelectorAll('input');
    inputs.forEach(input => {
        input.addEventListener('focus', () => {
            input.parentElement.classList.add('focused');
        });
        input.addEventListener('blur', () => {
            input.parentElement.classList.remove('focused');
        });
    });
});
