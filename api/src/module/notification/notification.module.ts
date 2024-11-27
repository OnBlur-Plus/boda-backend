import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { NotificationController } from './notification.controller';
import { NotificationService } from './notification.service';
import { Notification } from './entities/notification.entity';
import { User } from '../auth/entities/user.entity';
import { NotificationContent } from './entities/notification-content.entity';

@Module({
  imports: [TypeOrmModule.forFeature([Notification, NotificationContent, User])],
  controllers: [NotificationController],
  providers: [NotificationService],
  exports: [NotificationService],
})
export class NotificationModule {}
