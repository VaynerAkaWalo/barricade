export class AccountNotFoundError extends Error {}

export class WrongPasswordError extends Error {
	constructor(message = "Wrong password provided") {
		super(message);
	}
}

export class SessionExpiredError extends Error {
	constructor(message = "Session expired") {
		super(message);
	}
}

export class InternalServerError extends Error {
	constructor(message = "Internal server error") {
		super(message);
	}
}

export class FingerPrintMismach extends Error {
	constructor(message = "Unable to authenticate with provided session") {
		super(message);
	}
}
