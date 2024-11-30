import { Column, PrimaryGeneratedColumn } from 'typeorm';

export class Place {
  @PrimaryGeneratedColumn()
  id: number;

  @Column({ name: 'name' })
  name: string;
}
