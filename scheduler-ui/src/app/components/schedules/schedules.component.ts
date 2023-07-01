import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { Schedule } from 'src/app/models/schedule.model';
import { SchedulesService } from 'src/app/services/schedules.service';

@Component({
  selector: 'app-schedules',
  templateUrl: './schedules.component.html',
  styleUrls: ['./schedules.component.css']
})
export class SchedulesComponent {
  schedules: Schedule[] = [];

  constructor(private schedulesService: SchedulesService, private router: Router) { }

  ngOnInit(): void {
    this.loadSchedules();
  }

  loadSchedules() {
    this.schedulesService.getSchedules().subscribe(
      data => {
        this.schedules = data;
      }
    );
  }

  openSchedule(id: string): void {
    this.router.navigate([`/schedule/${id}`]);
  }
}
