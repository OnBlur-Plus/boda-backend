import { Controller, Get, ParseIntPipe, Query } from '@nestjs/common';
import { NotificationService } from './notification.service';

@Controller('notification')
export class NotificationController {
  constructor(private readonly notificationService: NotificationService) {}

  @Get()
  getAllNotifications(
    @Query('pageNum', ParseIntPipe) pageNum: number = 1,
    @Query('pageSize', ParseIntPipe) pageSize: number = 10,
  ) {}
}
