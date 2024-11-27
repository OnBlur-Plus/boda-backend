import { Body, Controller, DefaultValuePipe, Get, ParseIntPipe, Post, Query } from '@nestjs/common';
import { NotificationService } from './notification.service';
import { UserID } from './decorator/user-id.decorator';
import { RegisterDeviceTokenDto } from './dto/register-device-token.dto';

@Controller('notification')
export class NotificationController {
  constructor(private readonly notificationService: NotificationService) {}

  @Get()
  async getAllNotifications(
    @UserID() userId: number,
    @Query('pageNum', new DefaultValuePipe(1), ParseIntPipe) pageNum: number,
    @Query('pageSize', new DefaultValuePipe(10), ParseIntPipe) pageSize: number,
  ) {
    return await this.notificationService.getAllNotifications(userId, pageNum, pageSize);
  }

  @Post()
  async registerDeviceToken(@UserID() userId: number, @Body() registerDeviceTokenDto: RegisterDeviceTokenDto) {
    return await this.notificationService.registerDeviceToken(userId, registerDeviceTokenDto);
  }

  @Post('read')
  async readAllNotifications(@UserID() userId: number) {
    return await this.notificationService.readAllNotifications(userId, new Date());
  }
}
