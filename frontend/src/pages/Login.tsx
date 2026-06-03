import { Link } from "@tanstack/react-router";
import { Button } from "../components/Button/Button";
import { Input } from "../components/Input/Input";
import styles from "./Login.module.css";

export function LoginPage() {
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
				placeholder="Enter your password"
				autoComplete="current-password"
				required
			/>
			<Button type="submit" size="lg" className={styles.submit}>
				Sign In
			</Button>
			<div className={styles.links}>
				<Link to="/forgot-password" className={styles.link}>
					Forgot your password?
				</Link>
				<span className={styles.separator}>·</span>
				<Link to="/register" className={styles.link}>
					Create an account
				</Link>
			</div>
		</form>
	);
}
