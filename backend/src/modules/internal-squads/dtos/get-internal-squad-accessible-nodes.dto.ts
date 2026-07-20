import { createZodDto } from 'nestjs-zod';

import { GetInternalSquadAccessibleNodesCommand } from '@libs/contracts/commands';

export class GetInternalSquadAccessibleNodesRequestDto extends createZodDto(
    GetInternalSquadAccessibleNodesCommand.RequestSchema,
) {}
export class GetInternalSquadAccessibleNodesResponseDto extends createZodDto(
    GetInternalSquadAccessibleNodesCommand.ResponseSchema,
) {}
