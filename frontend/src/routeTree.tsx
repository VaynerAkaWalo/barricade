import {
  createRootRoute,
  createRoute,
} from '@tanstack/react-router'
import { RootLayout } from './components/RootLayout'
import { AuthLayout } from './components/AuthLayout/AuthLayout'
import { HomePage } from './pages/Home'
import { LoginPage } from './pages/Login'
import { RegisterPage } from './pages/Register'
import { ForgotPasswordPage } from './pages/ForgotPassword'

const rootRoute = createRootRoute({
  component: RootLayout,
})

const homeRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: HomePage,
})

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  component: () => (
    <AuthLayout
      title="Sign in to your account"
      subtitle="Enter your credentials to continue"
    >
      <LoginPage />
    </AuthLayout>
  ),
})

const registerRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/register',
  component: () => (
    <AuthLayout
      title="Create an account"
      subtitle="Get started with your identity provider"
    >
      <RegisterPage />
    </AuthLayout>
  ),
})

const forgotPasswordRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/forgot-password',
  component: () => (
    <AuthLayout
      title="Reset your password"
      subtitle="We will send you a reset link"
    >
      <ForgotPasswordPage />
    </AuthLayout>
  ),
})

export const routeTree = rootRoute.addChildren([
  homeRoute,
  loginRoute,
  registerRoute,
  forgotPasswordRoute,
])
