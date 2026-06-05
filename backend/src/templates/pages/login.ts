import { button } from "../components/button";
import { input } from "../components/input";

interface LoginFormProps {
	email?: string;
	error?: string;
}

export function loginForm(props: LoginFormProps): string {
	return `
<form method="POST" action="/login" class="form">
	${props.error ? `<p class="form-error" role="alert">${props.error}</p>` : ""}
	${input({
		name: "email",
		label: "Email",
		type: "email",
		placeholder: "you@example.com",
		autocomplete: "email",
		value: props.email,
		required: true,
	})}
	${input({
		name: "password",
		label: "Password",
		type: "password",
		placeholder: "Enter your password",
		autocomplete: "current-password",
		required: true,
	})}
	${button({ label: "Sign In", size: "lg", className: "submit" })}
	<div class="form-links">
		<a href="/register" class="form-link">Create an account</a>
	</div>
</form>`;
}

export function loginPage(props: LoginFormProps): string {
	return `
<div class="auth-page">
	<div class="auth-card card card-elevated card-pad-lg">
		<h1 class="auth-title">Welcome back</h1>
		<p class="auth-subtitle">Sign in to your account</p>
		${loginForm(props)}
	</div>
</div>`;
}
