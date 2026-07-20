import { createZodDto } from 'nestjs-zod';

import { GetConfigProfileByUuidCommand } from '@libs/contracts/commands';

export class GetConfigProfileByUuidRequestDto extends createZodDto(
    GetConfigProfileByUuidCommand.RequestSchema,
) {}
export class GetConfigProfileByUuidResponseDto extends createZodDto(
    GetConfigProfileByUuidCommand.ResponseSchema,
) {}
