import { escapeHtml } from "../layout";

interface InputProps {
	name: string;
	label: string;
	type?: string;
	value?: string;
	placeholder?: string;
	autocomplete?: string;
	error?: string;
	required?: boolean;
}

export function input({
	name,
	label,
	type = "text",
	value = "",
	placeholder,
	autocomplete,
	error,
	required,
}: InputProps): string {
	return `
<div class="input-wrapper">
	<label class="label" for="input-${name}">${escapeHtml(label)}</label>
	<input
		id="input-${name}"
		name="${escapeHtml(name)}"
		type="${escapeHtml(type)}"
		class="input${error ? " has-error" : ""}"
		value="${escapeHtml(value)}"
		${placeholder ? `placeholder="${escapeHtml(placeholder)}"` : ""}
		${autocomplete ? `autocomplete="${escapeHtml(autocomplete)}"` : ""}
		${required ? "required" : ""}
	>
	${error ? `<p class="input-error" role="alert">${escapeHtml(error)}</p>` : ""}
</div>`;
}
