import { createZodDto } from 'nestjs-zod';

import { DeleteInfraBillingNodeByUuidCommand } from '@libs/contracts/commands';

export class DeleteInfraBillingNodeByUuidRequestDto extends createZodDto(
    DeleteInfraBillingNodeByUuidCommand.RequestSchema,
) {}

export class DeleteInfraBillingNodeByUuidResponseDto extends createZodDto(
    DeleteInfraBillingNodeByUuidCommand.ResponseSchema,
) {}
