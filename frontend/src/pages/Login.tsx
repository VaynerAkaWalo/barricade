import { Link, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import { ApiError } from "../api/client";
import { useAuth } from "../auth/context";
import { Button } from "../components/Button/Button";
import { Input } from "../components/Input/Input";
import styles from "./Login.module.css";

export function LoginPage() {
	const router = useRouter();
	const { login } = useAuth();
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const [error, setError] = useState("");
	const [loading, setLoading] = useState(false);

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setError("");
		setLoading(true);
		try {
			await login(email, password);
			router.navigate({ to: "/dashboard" });
		} catch (err) {
			if (err instanceof ApiError && err.status === 500) {
				setError("Invalid email or password");
			} else if (err instanceof ApiError) {
				setError(err.message);
			} else {
				setError("An unexpected error occurred");
			}
		} finally {
			setLoading(false);
		}
	};

	return (
		<form className={styles.form} onSubmit={handleSubmit}>
			<Input
				label="Email"
				type="email"
				placeholder="you@example.com"
				autoComplete="email"
				required
				value={email}
				onChange={(e) => setEmail(e.target.value)}
			/>
			<Input
				label="Password"
				type="password"
				placeholder="Enter your password"
				autoComplete="current-password"
				required
				value={password}
				onChange={(e) => setPassword(e.target.value)}
			/>
			{error && (
				<p className={styles.error} role="alert">
					{error}
				</p>
			)}
			<Button
				type="submit"
				size="lg"
				className={styles.submit}
				loading={loading}
			>
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
