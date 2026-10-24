// resetPassword page texts
export default {
  title: 'Reset Password',
  subtitle: 'Set a new sign-in password',
  invalidToken: 'Invalid reset link: missing token parameter, please use the link from the email',
  placeholders: {
    newPassword: 'New password (at least 6 characters)',
    confirmPassword: 'Confirm new password'
  },
  rules: {
    newPasswordRequired: 'Please enter a new password',
    newPasswordMin: 'New password must be at least 6 characters',
    confirmPasswordRequired: 'Please enter the new password again',
    passwordMismatch: 'The two passwords do not match'
  },
  submit: 'Reset Password',
  backToLogin: 'Back to sign in',
  success: 'Password reset successfully, redirecting to sign in…',
  failed: 'Failed to reset password'
}
