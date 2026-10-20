// Keeps the encoded photos in .cache/images between deploys, so a deploy only
// encodes photos that are new or changed. Encoding is the slow part of the
// build; see internal/site/images.go. The build itself prunes entries it
// hasn't used in 30 days, so what's saved here doesn't grow without end.
const dir = ".cache/images";

module.exports = {
  async onPreBuild({ utils }) {
    if (await utils.cache.restore(dir)) {
      console.log(`Restored ${dir} from the build cache`);
    } else {
      console.log(`No ${dir} in the build cache yet; every photo will be encoded`);
    }
  },

  async onPostBuild({ utils }) {
    if (await utils.cache.save(dir)) {
      console.log(`Saved ${dir} to the build cache`);
    }
  },
};
