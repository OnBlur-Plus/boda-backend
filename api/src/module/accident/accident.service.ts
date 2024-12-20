import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Accident } from './entities/accident.entity';
import { Between, Repository } from 'typeorm';
import { UpdateAccidentDto } from './dto/update-accident.dto';
import { StartAccidentDto } from 'src/module/accident/dto/start-accident.dto';
import { StreamService } from 'src/module/stream/stream.service';
import { NotificationService } from '../notification/notification.service';
import { ACCIDENT_METADATA } from 'src/constant/accident';
import { EndAccidentDto } from 'src/module/accident/dto/end-accident.dto';

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

  async startAccident({ type, streamKey }: StartAccidentDto) {
    const startAt = new Date();
    const { reason, level } = ACCIDENT_METADATA[type];
    const stream = this.streamService.findStream(streamKey);

    if (!stream) {
      throw new NotFoundException('Not Found Stream');
    }

    const accident = this.accidentRepository.create({
      type,
      reason,
      startAt,
      level,
      streamKey,
    });

    await this.accidentRepository.save(accident);
    await this.notificationService.sendNotificationToAllUsers(accident);

    return {
      id: accident.id,
    };
  }

  async endAccident(endAccidentDto: EndAccidentDto) {
    return await this.accidentRepository.update(endAccidentDto.id, {
      endAt: Date(),
      videoUrl: endAccidentDto.videoUrl,
    });
  }

  async updateAccident(accidentId: number, updateaccidentDto: UpdateAccidentDto) {
    return await this.accidentRepository.update(accidentId, updateaccidentDto);
  }

  async deleteAccident(accidentId: number) {
    return await this.accidentRepository.delete(accidentId);
  }
}
