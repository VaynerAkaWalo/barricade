import { type ReactNode, useEffect, useState } from "react";
import { type Theme, ThemeContext } from "../theme/context";

function getInitialTheme(): Theme {
	if (typeof window !== "undefined") {
		const stored = localStorage.getItem("sk-theme");
		if (stored === "light" || stored === "dark") return stored;
		if (window.matchMedia("(prefers-color-scheme: dark)").matches)
			return "dark";
	}
	return "light";
}

export function ThemeProvider({ children }: { children: ReactNode }) {
	const [theme, setThemeState] = useState<Theme>(getInitialTheme);

	useEffect(() => {
		document.documentElement.setAttribute("data-theme", theme);
		localStorage.setItem("sk-theme", theme);
	}, [theme]);

	useEffect(() => {
		const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
		const handler = () => {
			const stored = localStorage.getItem("sk-theme");
			if (!stored) {
				setThemeState(mediaQuery.matches ? "dark" : "light");
			}
		};
		mediaQuery.addEventListener("change", handler);
		return () => mediaQuery.removeEventListener("change", handler);
	}, []);

	const toggleTheme = () => {
		setThemeState((prev) => (prev === "light" ? "dark" : "light"));
	};

	const setTheme = (t: Theme) => setThemeState(t);

	return (
		<ThemeContext.Provider value={{ theme, toggleTheme, setTheme }}>
			{children}
		</ThemeContext.Provider>
	);
}
