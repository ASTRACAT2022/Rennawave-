import { createZodDto } from 'nestjs-zod';

import { GetInboundsByProfileUuidCommand } from '@libs/contracts/commands';

export class GetInboundsByProfileUuidRequestDto extends createZodDto(
    GetInboundsByProfileUuidCommand.RequestSchema,
) {}
export class GetInboundsByProfileUuidResponseDto extends createZodDto(
    GetInboundsByProfileUuidCommand.ResponseSchema,
) {}
