import { createZodDto } from 'nestjs-zod';

import { DeleteUsersFromInternalSquadCommand } from '@libs/contracts/commands';

export class RemoveUsersFromInternalSquadRequestDto extends createZodDto(
    DeleteUsersFromInternalSquadCommand.RequestSchema,
) {}

export class RemoveUsersFromInternalSquadResponseDto extends createZodDto(
    DeleteUsersFromInternalSquadCommand.ResponseSchema,
) {}
