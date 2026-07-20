import { createZodDto } from 'nestjs-zod';

import { DeleteInternalSquadCommand } from '@libs/contracts/commands';

export class DeleteInternalSquadRequestDto extends createZodDto(
    DeleteInternalSquadCommand.RequestSchema,
) {}

export class DeleteInternalSquadResponseDto extends createZodDto(
    DeleteInternalSquadCommand.ResponseSchema,
) {}
