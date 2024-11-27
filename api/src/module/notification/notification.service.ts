import { Repository } from 'typeorm';
import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Notification } from './entities/notification.entity';

@Injectable()
export class NotificationService {
  constructor(@InjectRepository(Notification) private readonly notificationRepository: Repository<Notification>) {}

  async getAllNotifications(userId: number, pageNum: number, pageSize: number) {
    return await this.notificationRepository.find({
      where: { user: { id: userId } },
      skip: (pageNum - 1) * pageSize,
      take: pageSize,
    });
  }

  async readAllNotifications(userId: number, date: Date) {
    const { affected } = await this.notificationRepository.update(
      { user: { id: userId }, readedAt: null },
      { readedAt: date },
    );

    return affected;
  }
}
