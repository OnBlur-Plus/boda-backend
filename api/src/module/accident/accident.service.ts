import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Accident, AccidentLevel, AccidentType } from './entities/accident.entity';
import { Between, Repository } from 'typeorm';
import { UpdateAccidentDto } from './dto/update-accident.dto';
import { StartAccidentDto } from 'src/module/accident/dto/start-accident.dto';
import { StreamService } from 'src/module/stream/stream.service';
import { NotificationService } from '../notification/notification.service';

@Injectable()
export class AccidentService {
  constructor(
    @InjectRepository(Accident) private readonly accidentRepository: Repository<Accident>,
    private readonly streamService: StreamService,
    private readonly notificationService: NotificationService,
  ) {}

  async getAccidents(pageNum: number, pageSize: number) {
    return await this.accidentRepository.find({
      skip: (pageNum - 1) * pageSize,
      take: pageSize,
      order: { startAt: 'DESC' },
    });
  }

  async getAccidentsByStreamKey(streamKey: string) {
    return await this.accidentRepository.find({ where: { stream: { streamKey } } });
  }

  async getAccidentsByDate(date: Date) {
    const startDate = new Date(date.setUTCHours(0, 0, 0));
    const endDate = new Date(date.setUTCHours(23, 59, 59));

    return await this.accidentRepository.find({
      where: { startAt: Between(startDate, endDate) },
      order: { startAt: 'DESC' },
    });
  }

  async getAccidentWithStream(id: number) {
    return await this.accidentRepository.findOne({ where: { id }, relations: ['stream'] });
  }

  getAccidentLevel(type: AccidentType) {
    switch (type) {
      case AccidentType.NON_SAFETY_VEST:
      case AccidentType.NON_SAFETY_HELMET:
      case AccidentType.USE_PHONE_WHILE_WORKING:
        return AccidentLevel.LOW;
      case AccidentType.FALL:
        return AccidentLevel.MEDIUM;
      case AccidentType.SOS_REQUEST:
        return AccidentLevel.HIGH;
    }
  }

  async startAccident(startAccidentDto: StartAccidentDto) {
    const { type, reason, streamKey, startAt } = startAccidentDto;
    const stream = this.streamService.findStream(streamKey);

    if (!stream) {
      throw new NotFoundException('Not Found Stream');
    }

    const accident = this.accidentRepository.create({
      type,
      reason,
      startAt,
      level: this.getAccidentLevel(type),
      streamKey,
      createdAt: startAt,
    });

    await this.accidentRepository.save(accident);
    await this.notificationService.sendNotificationToAllUsers(accident);

    return accident;
  }

  async endAccident(id: number) {
    return await this.accidentRepository.update(id, { endAt: Date() });
  }

  async updateAccident(accidentId: number, updateaccidentDto: UpdateAccidentDto) {
    return await this.accidentRepository.update(accidentId, updateaccidentDto);
  }

  async deleteAccident(accidentId: number) {
    return await this.accidentRepository.delete(accidentId);
  }
}
