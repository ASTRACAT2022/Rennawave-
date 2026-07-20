import { createZodDto } from 'nestjs-zod';

import { GetInfraProviderByUuidCommand } from '@libs/contracts/commands';

export class GetInfraProviderByUuidRequestDto extends createZodDto(
    GetInfraProviderByUuidCommand.RequestSchema,
) {}

export class GetInfraProviderByUuidResponseDto extends createZodDto(
    GetInfraProviderByUuidCommand.ResponseSchema,
) {}
