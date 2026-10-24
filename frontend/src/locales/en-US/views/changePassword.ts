// changePassword page texts
export default {
  title: 'Change Password',
  subtitle: 'For security, you must change your password after first login (or after a password reset) before continuing to use the panel',
  placeholders: {
    oldPassword: 'Current password (see server startup log for the initial password)',
    newPassword: 'New password (at least 8 characters, letters and numbers)',
    confirmPassword: 'Confirm new password'
  },
  rules: {
    oldPasswordRequired: 'Please enter your current password',
    newPasswordRequired: 'Please enter a new password',
    newPasswordMin: 'New password must be at least 8 characters',
    newPasswordComplexity: 'New password must contain both letters and numbers',
    newPasswordSameAsOld: 'New password cannot be the same as the current password',
    confirmPasswordRequired: 'Please enter the new password again',
    passwordMismatch: 'The two passwords do not match'
  },
  submit: 'Confirm Change',
  success: 'Password changed successfully, please sign in with your new password',
  failed: 'Failed to change password'
}
