import { Link } from '@tanstack/react-router'
import { Button } from '../components/Button/Button'
import styles from './Home.module.css'

export function HomePage() {
  return (
    <div className={styles.page}>
      <div className={styles.content}>
        <div className={styles.badge}>OpenID Connect & OAuth 2.0</div>
        <h1 className={styles.title}>Barricade</h1>
        <p className={styles.subtitle}>
          Open-source identity provider. Secure, fast, and built for developers.
        </p>
        <div className={styles.actions}>
          <Link to="/login">
            <Button size="lg">Sign In</Button>
          </Link>
          <Link to="/register">
            <Button size="lg" variant="secondary">
              Create Account
            </Button>
          </Link>
        </div>
      </div>
    </div>
  )
}
