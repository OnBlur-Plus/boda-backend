import { Body, Controller, Post } from '@nestjs/common';
import { AuthService } from './auth.service';
import { VerifyDto } from './dto/verify.dto';
import { IsPublic } from './auth.guard';

@Controller('auth')
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  @IsPublic()
  @Post('verify')
  async verify(@Body() verifyDto: VerifyDto) {
    return await this.authService.verify(verifyDto);
  }
}
