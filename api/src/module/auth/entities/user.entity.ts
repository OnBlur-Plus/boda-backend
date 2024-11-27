import { Column, Entity, Index, PrimaryGeneratedColumn } from 'typeorm';

@Entity()
export class User {
  @PrimaryGeneratedColumn()
  id: number;

  @Column()
  @Index({ unique: true })
  pin: string;

  @Column()
  name: string;

  @Column({ nullable: true })
  deviceToken: string;
}
