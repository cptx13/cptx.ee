<template>
<AppLayout :pageTitle="'archive map ' + year + ' | cptx'" :description="'Heatmap of every croquis and recuerdo written in ' + year + '.'" :canonicalURL="'https://cptx.ee/addons/map/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons">
    <p class="addons-crumb"><a href="/addons/">addons</a> / archive map</p>
    <h1 class="addons-title">{{ year }}</h1>
    <p class="hm-stats">{{ total }} posts · {{ words }} words · {{ activeDays }} days · longest run {{ streakLabel }}</p>

    <div class="hm-filters">
      <p><span class="hm-filter-label">year</span><a v-for="o in years" :href="o.Href" :class="{ 'hm-opt': true, 'hm-active': o.Active }" rel="nofollow">{{ o.Label }}</a></p>
      <p><span class="hm-filter-label">section</span><a v-for="o in sections" :href="o.Href" :class="{ 'hm-opt': true, 'hm-active': o.Active }" rel="nofollow">{{ o.Label }}</a></p>
      <p><span class="hm-filter-label">tag</span><a v-for="o in tags" :href="o.Href" :class="{ 'hm-opt': true, 'hm-active': o.Active }" rel="nofollow">{{ o.Label }}</a></p>
      <p><span class="hm-filter-label">colour</span><a v-for="o in metrics" :href="o.Href" :class="{ 'hm-opt': true, 'hm-active': o.Active }" rel="nofollow">{{ o.Label }}</a></p>
    </div>

    <div class="hm-scroll">
      <div class="hm" role="grid" :aria-label="'Posts per day in ' + year">
        <div v-for="week in weeks" class="hm-week">
          <span class="hm-month">{{ week.Label }}</span>
          <template v-for="c in week.Days">
            <a v-if="c.Has" :class="c.Class" :href="c.Href" :title="c.Label" :aria-label="c.Label" rel="nofollow"></a>
            <span v-else :class="c.Class" :title="c.Label"></span>
          </template>
        </div>
      </div>
    </div>
    <p class="hm-legend">less <span class="hm-cell hm-l0"></span><span class="hm-cell hm-l1"></span><span class="hm-cell hm-l2"></span><span class="hm-cell hm-l3"></span><span class="hm-cell hm-l4"></span> more · by {{ metric }}</p>

    <div v-if="hasDayPosts" id="day" class="hm-day">
      <h2>{{ day }}</h2>
      <article v-for="p in dayPosts" class="post-entry">
        <span class="post-entry-date">{{ p.Section }}</span>
        <h2 class="post-entry-title"><a :href="p.Path">{{ p.Title }}</a></h2>
      </article>
    </div>

    <p class="hm-share">link to this view: <a :href="shareURL">{{ shareURL }}</a></p>
  </div>
</AppLayout>
</template>
