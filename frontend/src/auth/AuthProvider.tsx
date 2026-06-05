import { type ReactNode, useCallback, useEffect, useState } from "react";
import * as api from "../api/client";
import { AuthContext } from "./context";

export function AuthProvider({ children }: { children: ReactNode }) {
	const [isAuthenticated, setIsAuthenticated] = useState(false);
	const [isLoading, setIsLoading] = useState(true);

	useEffect(() => {
		api
			.authenticate()
			.then(() => setIsAuthenticated(true))
			.catch(() => setIsAuthenticated(false))
			.finally(() => setIsLoading(false));
	}, []);

	const login = useCallback(async (email: string, password: string) => {
		await api.login({ email, secret: password });
		setIsAuthenticated(true);
	}, []);

	const register = useCallback(
		async (email: string, password: string) => {
			await api.register({ email, secret: password });
			await login(email, password);
		},
		[login],
	);

	const logout = useCallback(async () => {
		await api.logout();
		setIsAuthenticated(false);
	}, []);

	return (
		<AuthContext.Provider
			value={{
				isAuthenticated,
				isLoading,
				login,
				register,
				logout,
			}}
		>
			{children}
		</AuthContext.Provider>
	);
}
