import { createZodDto } from 'nestjs-zod';

import { DeleteInfraBillingHistoryRecordCommand } from '@libs/contracts/commands';

export class DeleteInfraBillingHistoryRecordByUuidRequestDto extends createZodDto(
    DeleteInfraBillingHistoryRecordCommand.RequestSchema,
) {}

export class DeleteInfraBillingHistoryRecordByUuidResponseDto extends createZodDto(
    DeleteInfraBillingHistoryRecordCommand.ResponseSchema,
) {}
