import { createZodDto } from 'nestjs-zod';

import { UpdateInfraBillingNodeCommand } from '@libs/contracts/commands';

export class UpdateInfraBillingNodeRequestDto extends createZodDto(
    UpdateInfraBillingNodeCommand.RequestSchema,
) {}

export class UpdateInfraBillingNodeResponseDto extends createZodDto(
    UpdateInfraBillingNodeCommand.ResponseSchema,
) {}
