import { Link } from '@tanstack/react-router'
import { Button } from '../components/Button/Button'
import { Input } from '../components/Input/Input'
import styles from './Register.module.css'

export function RegisterPage() {
  return (
    <form className={styles.form} onSubmit={(e) => e.preventDefault()}>
      <Input
        label="Email"
        type="email"
        placeholder="you@example.com"
        autoComplete="email"
        required
      />
      <Input
        label="Password"
        type="password"
        placeholder="Create a password"
        autoComplete="new-password"
        required
      />
      <Input
        label="Confirm password"
        type="password"
        placeholder="Repeat your password"
        autoComplete="new-password"
        required
      />
      <Button type="submit" size="lg" className={styles.submit}>
        Create Account
      </Button>
      <div className={styles.links}>
        <span className={styles.text}>Already have an account?</span>
        <Link to="/login" className={styles.link}>
          Sign in
        </Link>
      </div>
    </form>
  )
}
