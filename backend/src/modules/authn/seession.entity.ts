export interface Session {
	id: string;
	owner: string;
	fingerPrint: bigint;
	createdAt: Date;
	expireAt: Date;
}
