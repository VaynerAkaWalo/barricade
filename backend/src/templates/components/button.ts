interface ButtonProps {
	label: string;
	type?: "submit" | "button";
	variant?: "primary" | "secondary" | "ghost" | "danger";
	size?: "sm" | "md" | "lg";
	className?: string;
}

export function button({
	label,
	type = "submit",
	variant = "primary",
	size = "md",
	className = "",
}: ButtonProps): string {
	return `<button type="${type}" class="btn btn-${variant} btn-${size} ${className}">${label}</button>`;
}
