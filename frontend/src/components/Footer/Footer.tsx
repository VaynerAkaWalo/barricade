import styles from './Footer.module.css'

export function Footer() {
  return (
    <footer className={styles.footer}>
      <div className={styles.inner}>
        <p className={styles.copyright}>&copy; {new Date().getFullYear()} Barricade. All rights reserved.</p>
      </div>
    </footer>
  )
}
