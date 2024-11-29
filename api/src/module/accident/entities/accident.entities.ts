import { Stream } from 'src/module/stream/entities/stream.entity';
import {
  Column,
  CreateDateColumn,
  Entity,
  JoinColumn,
  ManyToOne,
  PrimaryGeneratedColumn,
  UpdateDateColumn,
} from 'typeorm';

export enum AccidentType {
  NON_SAFETY_VEST = 'NON_SAFETY_VEST',
  NON_SAFETY_HELMET = 'NON_SAFETY_HELMET',
  FALL = 'FALL',
  USE_PHONE_WHILE_WORKING = 'USE_PHONE_WHILE_WORKING',
  SOS_REQUEST = 'SOS_REQUEST',
}

export enum AccidentLevel {
  LOW = 1,
  MEDIUM = 2,
  HIGH = 3,
}

@Entity()
export class Accident {
  @PrimaryGeneratedColumn()
  id: number;

  @JoinColumn({ name: 'stream_key' })
  @ManyToOne(() => Stream, { cascade: true })
  stream: Stream;

  @Column({ name: 'start_at' })
  startAt: Date;

  @Column({ name: 'end_at' })
  endAt: Date;

  @Column({ type: 'enum', enum: AccidentType })
  type: AccidentType;

  @Column({ type: 'enum', enum: AccidentLevel })
  level: AccidentLevel;

  @Column({ nullable: true })
  reason: string;

  @Column({ name: 'video_url', nullable: true })
  videoUrl: string;

  @CreateDateColumn({ name: 'created_at' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updated_at' })
  updatedAt: Date;
}
