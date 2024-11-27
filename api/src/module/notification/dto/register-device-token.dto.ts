import { IsString } from 'class-validator';

export class RegisterDeviceTokenDto {
  @IsString()
  deviceToken: string;
}
