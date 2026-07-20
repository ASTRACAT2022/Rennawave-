export class AddUserToNodeEvent {
    constructor(
        public readonly userUuid: string,
        public readonly prevVlessUuid?: string,
    ) {}
}
