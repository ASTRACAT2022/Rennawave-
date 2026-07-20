export class DeleteByUuidResponseModel {
    public readonly isDeleted: boolean;

    constructor(isDeleted: boolean) {
        this.isDeleted = isDeleted;
    }
}
