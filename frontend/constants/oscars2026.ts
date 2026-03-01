export interface OscarNominee {
  title: string;
  year: number;
  imdbId: string;
  mediaType: 'movie';
  personName?: string;
  country?: string;
}

export interface OscarCategory {
  id: string;
  name: string;
  icon: string;
  nominees: OscarNominee[];
}

export const OSCARS_2026_CATEGORIES: OscarCategory[] = [
  {
    id: 'best-picture',
    name: 'Best Picture',
    icon: 'trophy',
    nominees: [
      { title: 'Bugonia', year: 2025, imdbId: 'tt12300742', mediaType: 'movie' },
      { title: 'F1', year: 2025, imdbId: 'tt16311594', mediaType: 'movie' },
      { title: 'Frankenstein', year: 2025, imdbId: 'tt1312221', mediaType: 'movie' },
      { title: 'Hamnet', year: 2025, imdbId: 'tt14905854', mediaType: 'movie' },
      { title: 'Marty Supreme', year: 2025, imdbId: 'tt32916440', mediaType: 'movie' },
      { title: 'One Battle After Another', year: 2025, imdbId: 'tt30144839', mediaType: 'movie' },
      { title: 'The Secret Agent', year: 2025, imdbId: 'tt27847051', mediaType: 'movie' },
      { title: 'Sentimental Value', year: 2025, imdbId: 'tt27714581', mediaType: 'movie' },
      { title: 'Sinners', year: 2025, imdbId: 'tt31193180', mediaType: 'movie' },
      { title: 'Train Dreams', year: 2025, imdbId: 'tt29768334', mediaType: 'movie' },
    ],
  },
  {
    id: 'best-director',
    name: 'Best Director',
    icon: 'videocam',
    nominees: [
      { title: 'Hamnet', year: 2025, imdbId: 'tt14905854', mediaType: 'movie', personName: 'Chlo\u00e9 Zhao' },
      { title: 'Marty Supreme', year: 2025, imdbId: 'tt32916440', mediaType: 'movie', personName: 'Josh Safdie' },
      { title: 'One Battle After Another', year: 2025, imdbId: 'tt30144839', mediaType: 'movie', personName: 'Paul Thomas Anderson' },
      { title: 'Sentimental Value', year: 2025, imdbId: 'tt27714581', mediaType: 'movie', personName: 'Joachim Trier' },
      { title: 'Sinners', year: 2025, imdbId: 'tt31193180', mediaType: 'movie', personName: 'Ryan Coogler' },
    ],
  },
  {
    id: 'best-actor',
    name: 'Best Actor',
    icon: 'person',
    nominees: [
      { title: 'Marty Supreme', year: 2025, imdbId: 'tt32916440', mediaType: 'movie', personName: 'Timoth\u00e9e Chalamet' },
      { title: 'One Battle After Another', year: 2025, imdbId: 'tt30144839', mediaType: 'movie', personName: 'Leonardo DiCaprio' },
      { title: 'Blue Moon', year: 2025, imdbId: 'tt32536315', mediaType: 'movie', personName: 'Ethan Hawke' },
      { title: 'Sinners', year: 2025, imdbId: 'tt31193180', mediaType: 'movie', personName: 'Michael B. Jordan' },
      { title: 'The Secret Agent', year: 2025, imdbId: 'tt27847051', mediaType: 'movie', personName: 'Wagner Moura' },
    ],
  },
  {
    id: 'best-actress',
    name: 'Best Actress',
    icon: 'person',
    nominees: [
      { title: 'Hamnet', year: 2025, imdbId: 'tt14905854', mediaType: 'movie', personName: 'Jessie Buckley' },
      { title: 'If I Had Legs I\'d Kick You', year: 2025, imdbId: 'tt18382850', mediaType: 'movie', personName: 'Rose Byrne' },
      { title: 'Song Sung Blue', year: 2025, imdbId: 'tt30343021', mediaType: 'movie', personName: 'Kate Hudson' },
      { title: 'Sentimental Value', year: 2025, imdbId: 'tt27714581', mediaType: 'movie', personName: 'Renate Reinsve' },
      { title: 'Bugonia', year: 2025, imdbId: 'tt12300742', mediaType: 'movie', personName: 'Emma Stone' },
    ],
  },
  {
    id: 'best-supporting-actor',
    name: 'Best Supporting Actor',
    icon: 'people',
    nominees: [
      { title: 'One Battle After Another', year: 2025, imdbId: 'tt30144839', mediaType: 'movie', personName: 'Sean Penn' },
      { title: 'Frankenstein', year: 2025, imdbId: 'tt1312221', mediaType: 'movie', personName: 'Jacob Elordi' },
      { title: 'Sinners', year: 2025, imdbId: 'tt31193180', mediaType: 'movie', personName: 'Delroy Lindo' },
      { title: 'Sentimental Value', year: 2025, imdbId: 'tt27714581', mediaType: 'movie', personName: 'Stellan Skarsg\u00e5rd' },
      { title: 'One Battle After Another', year: 2025, imdbId: 'tt30144839', mediaType: 'movie', personName: 'Benicio del Toro' },
    ],
  },
  {
    id: 'best-supporting-actress',
    name: 'Best Supporting Actress',
    icon: 'people',
    nominees: [
      { title: 'Sentimental Value', year: 2025, imdbId: 'tt27714581', mediaType: 'movie', personName: 'Elle Fanning' },
      { title: 'Weapons', year: 2025, imdbId: 'tt26581740', mediaType: 'movie', personName: 'Amy Madigan' },
      { title: 'Sinners', year: 2025, imdbId: 'tt31193180', mediaType: 'movie', personName: 'Wunmi Mosaku' },
      { title: 'One Battle After Another', year: 2025, imdbId: 'tt30144839', mediaType: 'movie', personName: 'Teyana Taylor' },
      { title: 'Sentimental Value', year: 2025, imdbId: 'tt27714581', mediaType: 'movie', personName: 'Inga Ibsdotter Lilleaas' },
    ],
  },
  {
    id: 'best-original-screenplay',
    name: 'Best Original Screenplay',
    icon: 'document-text',
    nominees: [
      { title: 'Blue Moon', year: 2025, imdbId: 'tt32536315', mediaType: 'movie' },
      { title: 'It Was Just an Accident', year: 2025, imdbId: 'tt36491653', mediaType: 'movie' },
      { title: 'Marty Supreme', year: 2025, imdbId: 'tt32916440', mediaType: 'movie' },
      { title: 'Sentimental Value', year: 2025, imdbId: 'tt27714581', mediaType: 'movie' },
      { title: 'Sinners', year: 2025, imdbId: 'tt31193180', mediaType: 'movie' },
    ],
  },
  {
    id: 'best-adapted-screenplay',
    name: 'Best Adapted Screenplay',
    icon: 'document-text',
    nominees: [
      { title: 'Bugonia', year: 2025, imdbId: 'tt12300742', mediaType: 'movie' },
      { title: 'Frankenstein', year: 2025, imdbId: 'tt1312221', mediaType: 'movie' },
      { title: 'Hamnet', year: 2025, imdbId: 'tt14905854', mediaType: 'movie' },
      { title: 'One Battle After Another', year: 2025, imdbId: 'tt30144839', mediaType: 'movie' },
      { title: 'Train Dreams', year: 2025, imdbId: 'tt29768334', mediaType: 'movie' },
    ],
  },
  {
    id: 'best-animated-feature',
    name: 'Best Animated Feature Film',
    icon: 'color-palette',
    nominees: [
      { title: 'Arco', year: 2025, imdbId: 'tt14883538', mediaType: 'movie' },
      { title: 'Elio', year: 2025, imdbId: 'tt4900148', mediaType: 'movie' },
      { title: 'KPop Demon Hunters', year: 2025, imdbId: 'tt14205554', mediaType: 'movie' },
      { title: 'Little Am\u00e9lie or the Character of Rain', year: 2025, imdbId: 'tt29313285', mediaType: 'movie' },
      { title: 'Zootopia 2', year: 2025, imdbId: 'tt26443597', mediaType: 'movie' },
    ],
  },
  {
    id: 'best-international-feature',
    name: 'Best International Feature Film',
    icon: 'globe',
    nominees: [
      { title: 'The Secret Agent', year: 2025, imdbId: 'tt27847051', mediaType: 'movie', country: 'Brazil' },
      { title: 'It Was Just an Accident', year: 2025, imdbId: 'tt36491653', mediaType: 'movie', country: 'France' },
      { title: 'Sentimental Value', year: 2025, imdbId: 'tt27714581', mediaType: 'movie', country: 'Norway' },
      { title: 'Sir\u0101t', year: 2025, imdbId: 'tt32298285', mediaType: 'movie', country: 'Spain' },
      { title: 'The Voice of Hind Rajab', year: 2025, imdbId: 'tt36943034', mediaType: 'movie', country: 'Tunisia' },
    ],
  },
  {
    id: 'best-documentary-feature',
    name: 'Best Documentary Feature Film',
    icon: 'film',
    nominees: [
      { title: 'The Alabama Solution', year: 2025, imdbId: 'tt35307139', mediaType: 'movie' },
      { title: 'Come See Me in the Good Light', year: 2025, imdbId: 'tt34966013', mediaType: 'movie' },
      { title: 'Cutting Through Rocks', year: 2025, imdbId: 'tt10196414', mediaType: 'movie' },
      { title: 'Mr. Nobody Against Putin', year: 2025, imdbId: 'tt34965515', mediaType: 'movie' },
      { title: 'The Perfect Neighbor', year: 2025, imdbId: 'tt34962891', mediaType: 'movie' },
    ],
  },
  {
    id: 'best-documentary-short',
    name: 'Best Documentary Short Film',
    icon: 'film',
    nominees: [
      { title: 'All the Empty Rooms', year: 2025, imdbId: 'tt37798645', mediaType: 'movie' },
      { title: 'Armed Only with a Camera: The Life and Death of Brent Renaud', year: 2025, imdbId: 'tt35515733', mediaType: 'movie' },
      { title: 'Children No More: Were and Are Gone', year: 2025, imdbId: 'tt38691874', mediaType: 'movie' },
      { title: 'The Devil Is Busy', year: 2024, imdbId: 'tt34205548', mediaType: 'movie' },
      { title: 'Perfectly a Strangeness', year: 2024, imdbId: 'tt32205767', mediaType: 'movie' },
    ],
  },
  {
    id: 'best-live-action-short',
    name: 'Best Live Action Short Film',
    icon: 'film',
    nominees: [
      { title: 'Butcher\'s Stain', year: 2025, imdbId: 'tt39232815', mediaType: 'movie' },
      { title: 'A Friend of Dorothy', year: 2025, imdbId: 'tt35489516', mediaType: 'movie' },
      { title: 'Jane Austen\'s Period Drama', year: 2024, imdbId: 'tt31171760', mediaType: 'movie' },
      { title: 'The Singers', year: 2025, imdbId: 'tt33508491', mediaType: 'movie' },
      { title: 'Two People Exchanging Saliva', year: 2024, imdbId: 'tt33365259', mediaType: 'movie' },
    ],
  },
  {
    id: 'best-animated-short',
    name: 'Best Animated Short Film',
    icon: 'color-palette',
    nominees: [
      { title: 'Butterfly', year: 2024, imdbId: 'tt30970587', mediaType: 'movie' },
      { title: 'Forevergreen', year: 2025, imdbId: 'tt36454323', mediaType: 'movie' },
      { title: 'The Girl Who Cried Pearls', year: 2025, imdbId: 'tt36956820', mediaType: 'movie' },
      { title: 'Retirement Plan', year: 2024, imdbId: 'tt33019507', mediaType: 'movie' },
      { title: 'The Three Sisters', year: 2024, imdbId: 'tt35676079', mediaType: 'movie' },
    ],
  },
  {
    id: 'best-original-score',
    name: 'Best Original Score',
    icon: 'musical-notes',
    nominees: [
      { title: 'Bugonia', year: 2025, imdbId: 'tt12300742', mediaType: 'movie' },
      { title: 'Frankenstein', year: 2025, imdbId: 'tt1312221', mediaType: 'movie' },
      { title: 'Hamnet', year: 2025, imdbId: 'tt14905854', mediaType: 'movie' },
      { title: 'One Battle After Another', year: 2025, imdbId: 'tt30144839', mediaType: 'movie' },
      { title: 'Sinners', year: 2025, imdbId: 'tt31193180', mediaType: 'movie' },
    ],
  },
  {
    id: 'best-original-song',
    name: 'Best Original Song',
    icon: 'mic',
    nominees: [
      { title: 'Diane Warren: Relentless', year: 2024, imdbId: 'tt14588692', mediaType: 'movie', personName: '"Dear Me"' },
      { title: 'KPop Demon Hunters', year: 2025, imdbId: 'tt14205554', mediaType: 'movie', personName: '"Golden"' },
      { title: 'Sinners', year: 2025, imdbId: 'tt31193180', mediaType: 'movie', personName: '"I Lied to You"' },
      { title: 'Viva Verdi!', year: 2024, imdbId: 'tt3595454', mediaType: 'movie', personName: '"Sweet Dreams of Joy"' },
      { title: 'Train Dreams', year: 2025, imdbId: 'tt29768334', mediaType: 'movie', personName: '"Train Dreams"' },
    ],
  },
  {
    id: 'best-sound',
    name: 'Best Sound',
    icon: 'volume-high',
    nominees: [
      { title: 'F1', year: 2025, imdbId: 'tt16311594', mediaType: 'movie' },
      { title: 'Frankenstein', year: 2025, imdbId: 'tt1312221', mediaType: 'movie' },
      { title: 'One Battle After Another', year: 2025, imdbId: 'tt30144839', mediaType: 'movie' },
      { title: 'Sinners', year: 2025, imdbId: 'tt31193180', mediaType: 'movie' },
      { title: 'Sir\u0101t', year: 2025, imdbId: 'tt32298285', mediaType: 'movie' },
    ],
  },
  {
    id: 'best-production-design',
    name: 'Best Production Design',
    icon: 'construct',
    nominees: [
      { title: 'Frankenstein', year: 2025, imdbId: 'tt1312221', mediaType: 'movie' },
      { title: 'Hamnet', year: 2025, imdbId: 'tt14905854', mediaType: 'movie' },
      { title: 'Marty Supreme', year: 2025, imdbId: 'tt32916440', mediaType: 'movie' },
      { title: 'One Battle After Another', year: 2025, imdbId: 'tt30144839', mediaType: 'movie' },
      { title: 'Sinners', year: 2025, imdbId: 'tt31193180', mediaType: 'movie' },
    ],
  },
  {
    id: 'best-cinematography',
    name: 'Best Cinematography',
    icon: 'camera',
    nominees: [
      { title: 'Frankenstein', year: 2025, imdbId: 'tt1312221', mediaType: 'movie' },
      { title: 'Marty Supreme', year: 2025, imdbId: 'tt32916440', mediaType: 'movie' },
      { title: 'One Battle After Another', year: 2025, imdbId: 'tt30144839', mediaType: 'movie' },
      { title: 'Sinners', year: 2025, imdbId: 'tt31193180', mediaType: 'movie' },
      { title: 'Train Dreams', year: 2025, imdbId: 'tt29768334', mediaType: 'movie' },
    ],
  },
  {
    id: 'best-makeup-hairstyling',
    name: 'Best Makeup and Hairstyling',
    icon: 'brush',
    nominees: [
      { title: 'Frankenstein', year: 2025, imdbId: 'tt1312221', mediaType: 'movie' },
      { title: 'Kokuho', year: 2025, imdbId: 'tt35231039', mediaType: 'movie' },
      { title: 'Sinners', year: 2025, imdbId: 'tt31193180', mediaType: 'movie' },
      { title: 'The Smashing Machine', year: 2025, imdbId: 'tt11214558', mediaType: 'movie' },
      { title: 'The Ugly Stepsister', year: 2025, imdbId: 'tt29344903', mediaType: 'movie' },
    ],
  },
  {
    id: 'best-costume-design',
    name: 'Best Costume Design',
    icon: 'shirt',
    nominees: [
      { title: 'Avatar: Fire and Ash', year: 2025, imdbId: 'tt1757678', mediaType: 'movie' },
      { title: 'Frankenstein', year: 2025, imdbId: 'tt1312221', mediaType: 'movie' },
      { title: 'Hamnet', year: 2025, imdbId: 'tt14905854', mediaType: 'movie' },
      { title: 'Marty Supreme', year: 2025, imdbId: 'tt32916440', mediaType: 'movie' },
      { title: 'Sinners', year: 2025, imdbId: 'tt31193180', mediaType: 'movie' },
    ],
  },
  {
    id: 'best-film-editing',
    name: 'Best Film Editing',
    icon: 'cut',
    nominees: [
      { title: 'F1', year: 2025, imdbId: 'tt16311594', mediaType: 'movie' },
      { title: 'Marty Supreme', year: 2025, imdbId: 'tt32916440', mediaType: 'movie' },
      { title: 'One Battle After Another', year: 2025, imdbId: 'tt30144839', mediaType: 'movie' },
      { title: 'Sentimental Value', year: 2025, imdbId: 'tt27714581', mediaType: 'movie' },
      { title: 'Sinners', year: 2025, imdbId: 'tt31193180', mediaType: 'movie' },
    ],
  },
  {
    id: 'best-visual-effects',
    name: 'Best Visual Effects',
    icon: 'sparkles',
    nominees: [
      { title: 'Avatar: Fire and Ash', year: 2025, imdbId: 'tt1757678', mediaType: 'movie' },
      { title: 'F1', year: 2025, imdbId: 'tt16311594', mediaType: 'movie' },
      { title: 'Jurassic World Rebirth', year: 2025, imdbId: 'tt31036941', mediaType: 'movie' },
      { title: 'The Lost Bus', year: 2025, imdbId: 'tt21103218', mediaType: 'movie' },
      { title: 'Sinners', year: 2025, imdbId: 'tt31193180', mediaType: 'movie' },
    ],
  },
  {
    id: 'best-casting',
    name: 'Best Casting',
    icon: 'people-circle',
    nominees: [
      { title: 'Hamnet', year: 2025, imdbId: 'tt14905854', mediaType: 'movie' },
      { title: 'Marty Supreme', year: 2025, imdbId: 'tt32916440', mediaType: 'movie' },
      { title: 'One Battle After Another', year: 2025, imdbId: 'tt30144839', mediaType: 'movie' },
      { title: 'The Secret Agent', year: 2025, imdbId: 'tt27847051', mediaType: 'movie' },
      { title: 'Sinners', year: 2025, imdbId: 'tt31193180', mediaType: 'movie' },
    ],
  },
];
