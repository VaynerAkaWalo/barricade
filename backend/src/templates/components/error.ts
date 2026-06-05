import { escapeHtml } from "../layout";

export function errorBanner(message: string): string {
	return `<p class="form-error" role="alert">${escapeHtml(message)}</p>`;
}
