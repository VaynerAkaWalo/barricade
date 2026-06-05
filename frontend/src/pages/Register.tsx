import { Link, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import { ApiError } from "../api/client";
import { useAuth } from "../auth/context";
import { Button } from "../components/Button/Button";
import { Input } from "../components/Input/Input";
import styles from "./Register.module.css";

export function RegisterPage() {
	const router = useRouter();
	const { register } = useAuth();
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const [confirmPassword, setConfirmPassword] = useState("");
	const [error, setError] = useState("");
	const [loading, setLoading] = useState(false);

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setError("");

		if (password !== confirmPassword) {
			setError("Passwords do not match");
			return;
		}

		setLoading(true);
		try {
			await register(email, password);
			router.navigate({ to: "/dashboard" });
		} catch (err) {
			if (err instanceof ApiError && err.status === 422) {
				setError("Invalid email or password format");
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
				placeholder="Create a password"
				autoComplete="new-password"
				required
				value={password}
				onChange={(e) => setPassword(e.target.value)}
			/>
			<Input
				label="Confirm password"
				type="password"
				placeholder="Repeat your password"
				autoComplete="new-password"
				required
				value={confirmPassword}
				onChange={(e) => setConfirmPassword(e.target.value)}
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
				Create Account
			</Button>
			<div className={styles.links}>
				<span className={styles.text}>Already have an account?</span>
				<Link to="/login" className={styles.link}>
					Sign in
				</Link>
			</div>
		</form>
	);
}
