import { createZodDto } from 'nestjs-zod';

import { DeleteInfraProviderByUuidCommand } from '@libs/contracts/commands';

export class DeleteInfraProviderByUuidRequestDto extends createZodDto(
    DeleteInfraProviderByUuidCommand.RequestSchema,
) {}

export class DeleteInfraProviderByUuidResponseDto extends createZodDto(
    DeleteInfraProviderByUuidCommand.ResponseSchema,
) {}
