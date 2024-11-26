import { Controller, Get } from '@nestjs/common';
import { IsPublic } from '../auth/auth.guard';

@Controller('health')
export class HealthController {
  @IsPublic()
  @Get()
  check() {
    return 'OK';
  }
}
