import { Link } from '@tanstack/react-router'
import { Button } from '../components/Button/Button'
import { Input } from '../components/Input/Input'
import styles from './ForgotPassword.module.css'

export function ForgotPasswordPage() {
  return (
    <form className={styles.form} onSubmit={(e) => e.preventDefault()}>
      <Input
        label="Email"
        type="email"
        placeholder="you@example.com"
        autoComplete="email"
        required
      />
      <Button type="submit" size="lg" className={styles.submit}>
        Send Reset Link
      </Button>
      <div className={styles.links}>
        <Link to="/login" className={styles.link}>
          Back to sign in
        </Link>
      </div>
    </form>
  )
}
