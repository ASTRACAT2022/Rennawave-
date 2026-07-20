import { createZodDto } from 'nestjs-zod';

import { GetAllHostTagsCommand } from '@libs/contracts/commands';

export class GetAllHostTagsResponseDto extends createZodDto(GetAllHostTagsCommand.ResponseSchema) {}
