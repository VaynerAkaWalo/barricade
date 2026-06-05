import { button } from "../components/button";
import { input } from "../components/input";

interface RegisterFormProps {
	email?: string;
	error?: string;
}

export function registerForm(props: RegisterFormProps): string {
	return `
<form method="POST" action="/register" class="form">
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
		placeholder: "Create a password",
		autocomplete: "new-password",
		required: true,
	})}
	${input({
		name: "confirmPassword",
		label: "Confirm password",
		type: "password",
		placeholder: "Repeat your password",
		autocomplete: "new-password",
		required: true,
	})}
	${button({ label: "Create Account", size: "lg", className: "submit" })}
	<div class="form-links">
		<span class="form-text">Already have an account?</span>
		<a href="/login" class="form-link">Sign in</a>
	</div>
</form>`;
}

export function registerPage(props: RegisterFormProps): string {
	return `
<div class="auth-page">
	<div class="auth-card card card-elevated card-pad-lg">
		<h1 class="auth-title">Create an account</h1>
		<p class="auth-subtitle">Get started with Barricade</p>
		${registerForm(props)}
	</div>
</div>`;
}
