import { createZodDto } from 'nestjs-zod';

import { GetInternalSquadByUuidCommand } from '@libs/contracts/commands';

export class GetInternalSquadByUuidRequestDto extends createZodDto(
    GetInternalSquadByUuidCommand.RequestSchema,
) {}

export class GetInternalSquadByUuidResponseDto extends createZodDto(
    GetInternalSquadByUuidCommand.ResponseSchema,
) {}
