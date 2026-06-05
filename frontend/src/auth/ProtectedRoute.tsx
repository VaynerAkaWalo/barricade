import { useRouter } from "@tanstack/react-router";
import { type ReactNode, useEffect } from "react";
import { useAuth } from "./context";

export function ProtectedRoute({ children }: { children: ReactNode }) {
	const router = useRouter();
	const { isAuthenticated, isLoading } = useAuth();

	useEffect(() => {
		if (!isLoading && !isAuthenticated) {
			router.navigate({ to: "/login" });
		}
	}, [isAuthenticated, isLoading, router]);

	if (isLoading || !isAuthenticated) return null;

	return children;
}
