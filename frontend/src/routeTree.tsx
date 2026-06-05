import {
	createRootRoute,
	createRoute,
	useRouter,
} from "@tanstack/react-router";
import { useEffect } from "react";
import { useAuth } from "./auth/context";
import { ProtectedRoute } from "./auth/ProtectedRoute";
import { AuthLayout } from "./components/AuthLayout/AuthLayout";
import { RootLayout } from "./components/RootLayout";
import { DashboardPage } from "./pages/Dashboard";
import { ForgotPasswordPage } from "./pages/ForgotPassword";
import { LoginPage } from "./pages/Login";
import { RegisterPage } from "./pages/Register";

const rootRoute = createRootRoute({
	component: RootLayout,
});

function HomeRedirect() {
	const router = useRouter();
	const { isAuthenticated, isLoading } = useAuth();

	useEffect(() => {
		if (isLoading) return;
		router.navigate({ to: isAuthenticated ? "/dashboard" : "/login" });
	}, [isAuthenticated, isLoading, router]);

	return null;
}

const homeRoute = createRoute({
	getParentRoute: () => rootRoute,
	path: "/",
	component: HomeRedirect,
});

const loginRoute = createRoute({
	getParentRoute: () => rootRoute,
	path: "/login",
	component: () => (
		<AuthLayout
			title="Sign in to your account"
			subtitle="Enter your credentials to continue"
		>
			<LoginPage />
		</AuthLayout>
	),
});

const registerRoute = createRoute({
	getParentRoute: () => rootRoute,
	path: "/register",
	component: () => (
		<AuthLayout
			title="Create an account"
			subtitle="Get started with your identity provider"
		>
			<RegisterPage />
		</AuthLayout>
	),
});

const forgotPasswordRoute = createRoute({
	getParentRoute: () => rootRoute,
	path: "/forgot-password",
	component: () => (
		<AuthLayout
			title="Reset your password"
			subtitle="We will send you a reset link"
		>
			<ForgotPasswordPage />
		</AuthLayout>
	),
});

const dashboardRoute = createRoute({
	getParentRoute: () => rootRoute,
	path: "/dashboard",
	component: () => (
		<ProtectedRoute>
			<DashboardPage />
		</ProtectedRoute>
	),
});

export const routeTree = rootRoute.addChildren([
	homeRoute,
	loginRoute,
	registerRoute,
	forgotPasswordRoute,
	dashboardRoute,
]);
